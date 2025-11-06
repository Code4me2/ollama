# Chat Handler Refactor Design

## Core Requirements

1. **Multi-round tool execution**: Model can call tools multiple times
2. **Tool results feedback**: After executing tools, ALWAYS return results to model for further generation
3. **Clean exit**: When model generates response without tools, exit immediately
4. **Maximum rounds protection**: Prevent infinite loops with sensible limit
5. **Parallel/Sequential execution**: Use our new execution planning

## Current Flow Problems

The current nested loop structure is confusing because:
- Inner loop handles a single completion
- Outer loop handles multiple rounds
- Callback-based async pattern mixed with synchronous loop control
- Flags set in callback don't properly communicate with loop logic

## Proposed New Flow

```
START
  │
  ▼
┌─────────────────────────────┐
│  For each round (max 15)    │◄──────────┐
└─────────────────────────────┘           │
  │                                        │
  ▼                                        │
┌─────────────────────────────┐           │
│  Generate completion        │           │
│  (with current messages)    │           │
└─────────────────────────────┘           │
  │                                        │
  ▼                                        │
┌─────────────────────────────┐           │
│  Parse response             │           │
└─────────────────────────────┘           │
  │                                        │
  ▼                                        │
┌─────────────────────────────┐           │
│  Has tool calls?            │           │
└─────────────────────────────┘           │
  │                                        │
  ├─── NO: Exit loop ──────────────────►  END
  │                                        │
  ▼ YES                                    │
┌─────────────────────────────┐           │
│  Execute tools              │           │
│  (parallel/sequential)      │           │
└─────────────────────────────┘           │
  │                                        │
  ▼                                        │
┌─────────────────────────────┐           │
│  Add tool results to msgs   │           │
└─────────────────────────────┘           │
  │                                        │
  ▼                                        │
┌─────────────────────────────┐           │
│  Continue to next round     │───────────┘
│  (model processes results)  │
└─────────────────────────────┘
```

## Proposed Implementation Structure

```go
func (s *Server) ChatHandler(c *gin.Context) {
    // ... setup code ...
    
    go func() {
        defer close(ch)
        
        // Initialize messages
        currentMsgs := msgs
        
        // Multi-round execution loop
        for round := 0; round < maxRounds; round++ {
            slog.Debug("Starting round", "round", round)
            
            // Re-render prompt if not first round
            if round > 0 {
                prompt, images, err = chatPrompt(...)
                if err != nil {
                    ch <- gin.H{"error": err.Error()}
                    return
                }
            }
            
            // Execute completion and collect full response
            completionResponse, err := s.executeCompletionWithTools(
                c.Request.Context(),
                r, 
                prompt,
                images,
                opts,
                req,
                ch,  // for streaming
            )
            if err != nil {
                ch <- gin.H{"error": err.Error()}
                return
            }
            
            // Check if model called tools
            if len(completionResponse.ToolCalls) == 0 {
                // No tools called, conversation is complete
                slog.Debug("No tools called, completing conversation", "round", round)
                break
            }
            
            // Model called tools, execute them
            slog.Info("Executing tools", "count", len(completionResponse.ToolCalls), "round", round)
            
            // Analyze and execute tools
            executionPlan := mcpManager.AnalyzeExecutionPlan(completionResponse.ToolCalls)
            results := mcpManager.ExecuteWithPlan(completionResponse.ToolCalls, executionPlan)
            
            // Add assistant message with tool calls
            currentMsgs = append(currentMsgs, api.Message{
                Role:      "assistant",
                Content:   completionResponse.Content,
                ToolCalls: completionResponse.ToolCalls,
            })
            
            // Add tool results as individual messages
            for i, result := range results {
                toolMsg := api.Message{
                    Role:     "tool",
                    ToolName: completionResponse.ToolCalls[i].Function.Name,
                }
                
                if result.Error != nil {
                    toolMsg.Content = fmt.Sprintf("Error: %v", result.Error)
                } else {
                    toolMsg.Content = result.Content
                }
                
                currentMsgs = append(currentMsgs, toolMsg)
            }
            
            // Loop continues - model will process tool results in next round
        }
        
        // Check if we exhausted rounds
        if round >= maxRounds {
            slog.Warn("Maximum tool rounds reached", "rounds", maxRounds)
            ch <- gin.H{"error": "Maximum tool execution rounds exceeded"}
        }
    }()
    
    // ... response handling ...
}

// Helper function to execute completion and collect response
func (s *Server) executeCompletionWithTools(
    ctx context.Context,
    r *runner,
    prompt string,
    images []llm.ImageData,
    opts api.Options,
    req api.ChatRequest,
    ch chan any,
) (*CompletionResult, error) {
    result := &CompletionResult{}
    done := make(chan error, 1)
    
    err := r.Completion(ctx, llm.CompletionRequest{
        Prompt:   prompt,
        Images:   images,
        Format:   req.Format,
        Options:  opts,
        Shift:    req.Shift == nil || *req.Shift,
        Truncate: req.Truncate == nil || *req.Truncate,
    }, func(r llm.CompletionResponse) {
        // Build response
        res := api.ChatResponse{
            Model:     req.Model,
            CreatedAt: time.Now().UTC(),
            Message:   api.Message{Role: "assistant", Content: r.Content},
            Done:      r.Done,
            // ... metrics ...
        }
        
        // Parse tools if enabled
        if len(req.Tools) > 0 {
            toolCalls, content := toolParser.Add(res.Message.Content)
            if len(toolCalls) > 0 {
                result.ToolCalls = toolCalls
                res.Message.ToolCalls = toolCalls
                res.Message.Content = ""  // Clear content if tools detected
            } else {
                res.Message.Content = content
            }
        }
        
        // Stream response to client
        ch <- res
        
        // Store final result
        if r.Done {
            result.Content = res.Message.Content
            result.Done = true
            done <- nil
        }
    })
    
    if err != nil {
        return nil, err
    }
    
    // Wait for completion
    select {
    case err := <-done:
        return result, err
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}

type CompletionResult struct {
    Content   string
    ToolCalls []api.ToolCall
    Done      bool
}
```

## Key Benefits of This Design

1. **Clear Linear Flow**: Each round follows a simple sequence: generate → check tools → execute → continue or exit
2. **No Nested Loops**: Single loop with clear exit conditions
3. **Proper Tool Feedback**: Tool results always go back to model for processing
4. **Clean Exit**: When no tools called, exit immediately
5. **Synchronous Execution**: Helper function wraps async callback in synchronous interface
6. **Streaming Preserved**: Still sends responses to client as they arrive
7. **Error Handling**: Clear error propagation at each step

## Migration Plan

### Phase 1: Create Helper Function
1. Implement `executeCompletionWithTools` helper
2. Test it works with existing code
3. Ensure streaming still works

### Phase 2: Refactor Main Loop
1. Replace nested loops with single loop
2. Use helper function for completion
3. Implement clear tool detection and execution

### Phase 3: Testing
1. Test simple responses (no tools)
2. Test single tool execution
3. Test multi-round tool execution
4. Test error cases
5. Test maximum rounds protection

## Edge Cases to Handle

1. **Model calls same tool repeatedly**: Should work, up to max rounds
2. **Tool execution fails**: Error returned to model, can retry or explain
3. **Model generates both content and tools**: Handle appropriately
4. **Context cancellation**: Clean shutdown at any point
5. **Streaming interruption**: Proper cleanup

## Success Criteria

1. ✅ "Hello" responds immediately without loops
2. ✅ Tool execution works correctly
3. ✅ Tool results prompt further generation
4. ✅ Multiple rounds of tools work
5. ✅ No "Maximum rounds exceeded" for normal conversations
6. ✅ Parallel/Sequential execution planning works
7. ✅ Streaming responses still work
8. ✅ Error handling is robust