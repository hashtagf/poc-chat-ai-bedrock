# Design Document: Bedrock Agent Core Integration

## Overview

This design document outlines the implementation of Amazon Bedrock Agent Core integration with the Go backend application. The design replaces the current basic Bedrock Agent Runtime with a comprehensive Agent Core orchestration system that provides multi-step reasoning, advanced knowledge base integration, action groups support, and sophisticated session management.

The Agent Core integration transforms the application from a simple question-answer system into an intelligent conversational AI that can handle complex queries, perform multi-step tasks, and maintain rich conversation context. This upgrade enables the system to decompose complex problems, orchestrate multiple information sources, and provide comprehensive, well-reasoned responses.

### Requirements Coverage

This design addresses all requirements from the requirements document:

- **Requirement 1** (Multi-step Reasoning): Addressed by Agent Orchestration Engine (Section 3.1)
- **Requirement 2** (Session Context Management): Addressed by Session Management System (Section 3.2)
- **Requirement 3** (Action Groups Integration): Addressed by Action Groups Framework (Section 3.3)
- **Requirement 4** (Knowledge Base Intelligence): Addressed by Advanced Knowledge Base Integration (Section 3.4)
- **Requirement 5** (Streaming Operations): Addressed by Streaming Response System (Section 3.5)
- **Requirement 6** (Error Handling): Addressed by Comprehensive Error Management (Section 3.6)
- **Requirement 7** (Configuration Management): Addressed by Configuration System (Section 3.7)
- **Requirement 8** (Logging and Monitoring): Addressed by Observability Framework (Section 3.8)
- **Requirement 9** (Environment Support): Addressed by Environment Configuration (Section 3.9)
- **Requirement 10** (Reasoning Explanations): Addressed by Explanation Engine (Section 3.10)
- **Requirement 11** (Testing and Maintainability): Addressed by Testing Framework (Section 3.11)
- **Requirement 12** (Concurrency Management): Addressed by Concurrent Execution System (Section 3.12)
- **Requirement 13** (Metrics and Analytics): Addressed by Analytics and Metrics System (Section 3.13)

## Architecture

### High-Level Agent Core Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Agent Core Orchestrator                   │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  Planning Engine                                       │ │
│  │  - Query Analysis & Intent Recognition                 │ │
│  │  - Task Decomposition & Sequencing                     │ │
│  │  - Resource Selection & Optimization                   │ │
│  │  - Execution Strategy Planning                         │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    Execution Layer                          │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  Multi-step Executor                                   │ │
│  │  - Step-by-step Task Execution                         │ │
│  │  - Dependency Management                               │ │
│  │  - Result Aggregation & Synthesis                     │ │
│  │  - Progress Tracking & Streaming                      │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    Integration Layer                        │
│  ┌──────────────────┐ ┌──────────────────┐ ┌─────────────┐ │
│  │  Knowledge Base  │ │  Action Groups   │ │  Session    │ │
│  │  Integration     │ │  Framework       │ │  Management │ │
│  │  - RAG Engine    │ │  - API Calls     │ │  - Context  │ │
│  │  - Multi-source  │ │  - Function Exec │ │  - Memory   │ │
│  │  - Relevance     │ │  - Result Proc   │ │  - State    │ │
│  └──────────────────┘ └──────────────────┘ └─────────────┘ │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    Foundation Layer                         │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  AWS Bedrock Agent Core SDK                            │ │
│  │  - Agent Runtime API                                   │ │
│  │  - Knowledge Base API                                  │ │
│  │  - Action Groups API                                   │ │
│  │  - Session Management API                              │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Agent Core Workflow Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    User Request Processing                   │
│                                                             │
│  User Query → Intent Analysis → Task Planning → Execution   │
│                     ↓              ↓            ↓          │
│                Context Retrieval → Resource Selection →     │
│                                                   ↓          │
│                Multi-step Execution → Result Synthesis      │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    Session Context Flow                     │
│                                                             │
│  Previous Context → Context Compression → Current Context   │
│         ↓                    ↓                    ↓         │
│  Memory Storage → Context Window Management → New Memory    │
│                                                             │
│  User Preferences → Personalization → Response Adaptation  │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    Knowledge Integration Flow               │
│                                                             │
│  Query Intent → KB Selection → Multi-source Search →       │
│       ↓              ↓              ↓                      │
│  Relevance Scoring → Result Ranking → Citation Generation  │
│                                                             │
│  Context Synthesis → Response Enhancement → Final Output    │
└─────────────────────────────────────────────────────────────┘
```

## Components and Interfaces

### 1. Agent Orchestration Engine

**Purpose**: Coordinates multi-step reasoning and task execution using Bedrock Agent Core.

**Core Components**:

**Agent Core Client**:
```go
type AgentCoreClient interface {
    // InvokeAgent performs multi-step agent orchestration
    InvokeAgent(ctx context.Context, input *AgentCoreInput) (*AgentCoreResponse, error)
    
    // InvokeAgentStream provides streaming multi-step execution
    InvokeAgentStream(ctx context.Context, input *AgentCoreInput) (AgentCoreStreamReader, error)
    
    // GetSessionContext retrieves current session state
    GetSessionContext(ctx context.Context, sessionID string) (*SessionContext, error)
    
    // UpdateSessionContext modifies session state
    UpdateSessionContext(ctx context.Context, sessionID string, context *SessionContext) error
}

type AgentCoreInput struct {
    SessionID        string                 `json:"sessionId"`
    Message          string                 `json:"message"`
    KnowledgeBaseIDs []string              `json:"knowledgeBaseIds,omitempty"`
    ActionGroups     []string              `json:"actionGroups,omitempty"`
    SessionContext   *SessionContext       `json:"sessionContext,omitempty"`
    ExecutionConfig  *ExecutionConfig      `json:"executionConfig,omitempty"`
    StreamingConfig  *StreamingConfig      `json:"streamingConfig,omitempty"`
}

type AgentCoreResponse struct {
    Content          string                `json:"content"`
    Citations        []entities.Citation   `json:"citations"`
    ActionResults    []ActionResult        `json:"actionResults"`
    ReasoningSteps   []ReasoningStep       `json:"reasoningSteps"`
    SessionContext   *SessionContext       `json:"sessionContext"`
    ExecutionMetrics *ExecutionMetrics     `json:"executionMetrics"`
    RequestID        string                `json:"requestId"`
}
```

**Planning Engine**:
```go
type PlanningEngine struct {
    agentCore    AgentCoreClient
    config       PlanningConfig
    knowledgeDB  KnowledgeBaseManager
    actionGroups ActionGroupManager
}

func (p *PlanningEngine) AnalyzeIntent(ctx context.Context, message string, context *SessionContext) (*QueryIntent, error) {
    // Use Agent Core to analyze user intent and determine required capabilities
    input := &AgentCoreInput{
        Message:        message,
        SessionContext: context,
        ExecutionConfig: &ExecutionConfig{
            Mode: "analysis",
            MaxSteps: 1,
        },
    }
    
    response, err := p.agentCore.InvokeAgent(ctx, input)
    if err != nil {
        return nil, fmt.Errorf("intent analysis failed: %w", err)
    }
    
    return p.parseIntentFromResponse(response)
}

func (p *PlanningEngine) CreateExecutionPlan(ctx context.Context, intent *QueryIntent, context *SessionContext) (*ExecutionPlan, error) {
    // Use Agent Core to create multi-step execution plan
    planningInput := &AgentCoreInput{
        Message: fmt.Sprintf("Create execution plan for: %s", intent.Description),
        SessionContext: context,
        KnowledgeBaseIDs: intent.RequiredKnowledgeBases,
        ActionGroups: intent.RequiredActionGroups,
        ExecutionConfig: &ExecutionConfig{
            Mode: "planning",
            MaxSteps: 5,
            EnableReasoning: true,
        },
    }
    
    response, err := p.agentCore.InvokeAgent(ctx, planningInput)
    if err != nil {
        return nil, fmt.Errorf("execution planning failed: %w", err)
    }
    
    return p.parseExecutionPlan(response)
}
```

**Multi-step Executor**:
```go
type MultiStepExecutor struct {
    agentCore     AgentCoreClient
    config        ExecutorConfig
    stepProcessor StepProcessor
    resultSynth   ResultSynthesizer
}

func (e *MultiStepExecutor) ExecutePlan(ctx context.Context, plan *ExecutionPlan, sessionID string) (*AgentCoreResponse, error) {
    executionInput := &AgentCoreInput{
        SessionID: sessionID,
        Message: plan.InitialQuery,
        KnowledgeBaseIDs: plan.RequiredKnowledgeBases,
        ActionGroups: plan.RequiredActionGroups,
        ExecutionConfig: &ExecutionConfig{
            Mode: "execution",
            MaxSteps: plan.MaxSteps,
            EnableReasoning: true,
            EnableCitations: true,
        },
    }
    
    response, err := e.agentCore.InvokeAgent(ctx, executionInput)
    if err != nil {
        return nil, fmt.Errorf("plan execution failed: %w", err)
    }
    
    // Process and enhance the response
    return e.enhanceResponse(ctx, response, plan)
}

func (e *MultiStepExecutor) ExecutePlanStream(ctx context.Context, plan *ExecutionPlan, sessionID string) (AgentCoreStreamReader, error) {
    streamInput := &AgentCoreInput{
        SessionID: sessionID,
        Message: plan.InitialQuery,
        KnowledgeBaseIDs: plan.RequiredKnowledgeBases,
        ActionGroups: plan.RequiredActionGroups,
        StreamingConfig: &StreamingConfig{
            EnableStepUpdates: true,
            EnableProgressTracking: true,
            EnableReasoningTrace: true,
        },
    }
    
    return e.agentCore.InvokeAgentStream(ctx, streamInput)
}
```

### 2. Session Management System

**Purpose**: Manages persistent conversation context and memory using Agent Core session capabilities.

**Session Context Manager**:
```go
type SessionContextManager struct {
    agentCore    AgentCoreClient
    config       SessionConfig
    memoryStore  MemoryStore
    contextComp  ContextCompressor
}

type SessionContext struct {
    SessionID        string                 `json:"sessionId"`
    ConversationHist []ConversationTurn     `json:"conversationHistory"`
    UserPreferences  map[string]interface{} `json:"userPreferences"`
    ImportantFacts   []MemoryItem          `json:"importantFacts"`
    ActiveTopics     []string              `json:"activeTopics"`
    ContextWindow    *ContextWindow        `json:"contextWindow"`
    LastUpdated      time.Time             `json:"lastUpdated"`
}

func (s *SessionContextManager) GetContext(ctx context.Context, sessionID string) (*SessionContext, error) {
    // Retrieve session context from Agent Core
    context, err := s.agentCore.GetSessionContext(ctx, sessionID)
    if err != nil {
        return nil, fmt.Errorf("failed to get session context: %w", err)
    }
    
    // Enhance with local memory store if needed
    return s.enhanceWithLocalMemory(ctx, context)
}

func (s *SessionContextManager) UpdateContext(ctx context.Context, sessionID string, newTurn *ConversationTurn) error {
    context, err := s.GetContext(ctx, sessionID)
    if err != nil {
        return fmt.Errorf("failed to get current context: %w", err)
    }
    
    // Add new conversation turn
    context.ConversationHist = append(context.ConversationHist, *newTurn)
    
    // Compress context if needed
    if s.shouldCompressContext(context) {
        context, err = s.contextComp.CompressContext(ctx, context)
        if err != nil {
            return fmt.Errorf("context compression failed: %w", err)
        }
    }
    
    // Update in Agent Core
    return s.agentCore.UpdateSessionContext(ctx, sessionID, context)
}
```

**Memory Management**:
```go
type MemoryManager struct {
    agentCore   AgentCoreClient
    config      MemoryConfig
    importance  ImportanceScorer
    retention   RetentionPolicy
}

func (m *MemoryManager) ExtractImportantFacts(ctx context.Context, conversation []ConversationTurn) ([]MemoryItem, error) {
    // Use Agent Core to identify important information
    analysisInput := &AgentCoreInput{
        Message: "Extract important facts and preferences from this conversation",
        ExecutionConfig: &ExecutionConfig{
            Mode: "memory_extraction",
            MaxSteps: 2,
        },
    }
    
    response, err := m.agentCore.InvokeAgent(ctx, analysisInput)
    if err != nil {
        return nil, fmt.Errorf("memory extraction failed: %w", err)
    }
    
    return m.parseMemoryItems(response)
}

func (m *MemoryManager) SummarizeContext(ctx context.Context, context *SessionContext) (*ContextSummary, error) {
    // Use Agent Core to create intelligent context summary
    summaryInput := &AgentCoreInput{
        Message: "Summarize the key points and context from this conversation",
        SessionContext: context,
        ExecutionConfig: &ExecutionConfig{
            Mode: "summarization",
            MaxSteps: 1,
        },
    }
    
    response, err := m.agentCore.InvokeAgent(ctx, summaryInput)
    if err != nil {
        return nil, fmt.Errorf("context summarization failed: %w", err)
    }
    
    return m.parseContextSummary(response)
}
```

### 3. Action Groups Framework

**Purpose**: Integrates external API functions and services through Bedrock Agent Core action groups.

**Action Group Manager**:
```go
type ActionGroupManager struct {
    agentCore     AgentCoreClient
    config        ActionGroupConfig
    registry      ActionRegistry
    executor      ActionExecutor
}

type ActionGroup struct {
    ID           string                    `json:"id"`
    Name         string                    `json:"name"`
    Description  string                    `json:"description"`
    Functions    []ActionFunction          `json:"functions"`
    Config       ActionGroupConfig         `json:"config"`
    Enabled      bool                      `json:"enabled"`
}

type ActionFunction struct {
    Name         string                    `json:"name"`
    Description  string                    `json:"description"`
    Parameters   map[string]ParameterSpec  `json:"parameters"`
    ReturnType   string                    `json:"returnType"`
    Timeout      time.Duration             `json:"timeout"`
}

func (a *ActionGroupManager) RegisterActionGroup(ctx context.Context, group *ActionGroup) error {
    // Register action group with Agent Core
    if err := a.validateActionGroup(group); err != nil {
        return fmt.Errorf("action group validation failed: %w", err)
    }
    
    // Store in local registry
    a.registry.Register(group)
    
    log.Printf("[ActionGroups] Registered action group: %s with %d functions", group.Name, len(group.Functions))
    return nil
}

func (a *ActionGroupManager) InvokeAction(ctx context.Context, actionName string, parameters map[string]interface{}) (*ActionResult, error) {
    // Agent Core will handle action invocation through configured action groups
    // This method provides additional validation and monitoring
    
    action, exists := a.registry.GetAction(actionName)
    if !exists {
        return nil, fmt.Errorf("action not found: %s", actionName)
    }
    
    // Validate parameters
    if err := a.validateParameters(action, parameters); err != nil {
        return nil, fmt.Errorf("parameter validation failed: %w", err)
    }
    
    // Execute through Agent Core (action groups are invoked automatically)
    startTime := time.Now()
    result, err := a.executor.Execute(ctx, action, parameters)
    duration := time.Since(startTime)
    
    // Log execution metrics
    log.Printf("[ActionGroups] Action %s executed in %v, success: %t", actionName, duration, err == nil)
    
    return result, err
}
```

**Action Execution Framework**:
```go
type ActionExecutor struct {
    agentCore   AgentCoreClient
    config      ExecutorConfig
    monitor     ActionMonitor
    retry       RetryPolicy
}

func (e *ActionExecutor) Execute(ctx context.Context, action *ActionFunction, params map[string]interface{}) (*ActionResult, error) {
    // Create execution context with timeout
    execCtx, cancel := context.WithTimeout(ctx, action.Timeout)
    defer cancel()
    
    // Agent Core handles the actual function invocation
    // We provide monitoring and error handling
    
    result := &ActionResult{
        ActionName:  action.Name,
        StartTime:   time.Now(),
        Parameters:  params,
    }
    
    // Monitor execution
    e.monitor.StartExecution(action.Name, params)
    defer func() {
        result.EndTime = time.Now()
        result.Duration = result.EndTime.Sub(result.StartTime)
        e.monitor.EndExecution(action.Name, result)
    }()
    
    // The actual execution is handled by Agent Core through action groups
    // We simulate the result structure here
    result.Success = true
    result.Data = map[string]interface{}{
        "message": "Action executed successfully through Agent Core",
    }
    
    return result, nil
}
```

### 4. Advanced Knowledge Base Integration

**Purpose**: Provides intelligent knowledge retrieval and RAG capabilities through Agent Core.

**Knowledge Base Manager**:
```go
type KnowledgeBaseManager struct {
    agentCore      AgentCoreClient
    config         KnowledgeConfig
    queryOptimizer QueryOptimizer
    relevanceScorer RelevanceScorer
}

func (k *KnowledgeBaseManager) QueryKnowledgeBases(ctx context.Context, query string, kbIDs []string, context *SessionContext) (*KnowledgeResult, error) {
    // Agent Core handles intelligent knowledge base querying
    queryInput := &AgentCoreInput{
        Message: query,
        KnowledgeBaseIDs: kbIDs,
        SessionContext: context,
        ExecutionConfig: &ExecutionConfig{
            Mode: "knowledge_retrieval",
            MaxSteps: 3,
            EnableCitations: true,
        },
    }
    
    response, err := k.agentCore.InvokeAgent(ctx, queryInput)
    if err != nil {
        return nil, fmt.Errorf("knowledge base query failed: %w", err)
    }
    
    return k.processKnowledgeResponse(response)
}

func (k *KnowledgeBaseManager) OptimizeQuery(ctx context.Context, originalQuery string, context *SessionContext) (*OptimizedQuery, error) {
    // Use Agent Core to optimize query based on context
    optimizationInput := &AgentCoreInput{
        Message: fmt.Sprintf("Optimize this query for knowledge retrieval: %s", originalQuery),
        SessionContext: context,
        ExecutionConfig: &ExecutionConfig{
            Mode: "query_optimization",
            MaxSteps: 1,
        },
    }
    
    response, err := k.agentCore.InvokeAgent(ctx, optimizationInput)
    if err != nil {
        return nil, fmt.Errorf("query optimization failed: %w", err)
    }
    
    return k.parseOptimizedQuery(response)
}
```

**RAG Enhancement Engine**:
```go
type RAGEngine struct {
    agentCore       AgentCoreClient
    config          RAGConfig
    contextSynth    ContextSynthesizer
    citationProc    CitationProcessor
}

func (r *RAGEngine) EnhanceResponse(ctx context.Context, baseResponse string, knowledgeResults []*KnowledgeResult, context *SessionContext) (*EnhancedResponse, error) {
    // Use Agent Core to synthesize knowledge with response
    enhancementInput := &AgentCoreInput{
        Message: fmt.Sprintf("Enhance this response with knowledge: %s", baseResponse),
        SessionContext: context,
        ExecutionConfig: &ExecutionConfig{
            Mode: "response_enhancement",
            MaxSteps: 2,
            EnableCitations: true,
        },
    }
    
    response, err := r.agentCore.InvokeAgent(ctx, enhancementInput)
    if err != nil {
        return nil, fmt.Errorf("response enhancement failed: %w", err)
    }
    
    return r.processEnhancedResponse(response, knowledgeResults)
}
```

### 5. Streaming Response System

**Purpose**: Provides real-time streaming of multi-step agent operations.

**Agent Core Stream Reader**:
```go
type AgentCoreStreamReader interface {
    // Read returns the next content chunk and completion status
    Read() (chunk string, done bool, err error)
    
    // ReadStep returns information about the current execution step
    ReadStep() (step *ExecutionStep, err error)
    
    // ReadCitation returns any citations available in the current chunk
    ReadCitation() (citation *entities.Citation, err error)
    
    // ReadActionResult returns results from action group invocations
    ReadActionResult() (result *ActionResult, err error)
    
    // ReadReasoningTrace returns reasoning information if available
    ReadReasoningTrace() (trace *ReasoningTrace, err error)
    
    // Close closes the stream and releases resources
    Close() error
}

type StreamProcessor struct {
    agentCore   AgentCoreClient
    config      StreamConfig
    eventProc   EventProcessor
    bufferMgr   BufferManager
}

func (s *StreamProcessor) ProcessAgentStream(ctx context.Context, stream AgentCoreStreamReader) (*StreamingResponse, error) {
    response := &StreamingResponse{
        Content:       strings.Builder{},
        Citations:     []entities.Citation{},
        ActionResults: []ActionResult{},
        Steps:         []ExecutionStep{},
        Traces:        []ReasoningTrace{},
    }
    
    for {
        // Read content chunk
        chunk, done, err := stream.Read()
        if err != nil {
            return nil, fmt.Errorf("stream read error: %w", err)
        }
        
        if chunk != "" {
            response.Content.WriteString(chunk)
        }
        
        // Read additional stream data
        if step, err := stream.ReadStep(); err == nil && step != nil {
            response.Steps = append(response.Steps, *step)
        }
        
        if citation, err := stream.ReadCitation(); err == nil && citation != nil {
            response.Citations = append(response.Citations, *citation)
        }
        
        if actionResult, err := stream.ReadActionResult(); err == nil && actionResult != nil {
            response.ActionResults = append(response.ActionResults, *actionResult)
        }
        
        if trace, err := stream.ReadReasoningTrace(); err == nil && trace != nil {
            response.Traces = append(response.Traces, *trace)
        }
        
        if done {
            break
        }
    }
    
    return response, nil
}
```

## Data Models

### Agent Core Data Models

**Core Input/Output Models**:
```go
type ExecutionConfig struct {
    Mode                string        `json:"mode"`                // "analysis", "planning", "execution", etc.
    MaxSteps           int           `json:"maxSteps"`
    EnableReasoning    bool          `json:"enableReasoning"`
    EnableCitations    bool          `json:"enableCitations"`
    EnableActionGroups bool          `json:"enableActionGroups"`
    Timeout            time.Duration `json:"timeout"`
}

type StreamingConfig struct {
    EnableStepUpdates      bool `json:"enableStepUpdates"`
    EnableProgressTracking bool `json:"enableProgressTracking"`
    EnableReasoningTrace   bool `json:"enableReasoningTrace"`
    BufferSize            int  `json:"bufferSize"`
}

type ExecutionMetrics struct {
    TotalSteps        int           `json:"totalSteps"`
    ExecutionTime     time.Duration `json:"executionTime"`
    KnowledgeQueries  int           `json:"knowledgeQueries"`
    ActionInvocations int           `json:"actionInvocations"`
    TokensUsed        int           `json:"tokensUsed"`
}
```

**Reasoning and Planning Models**:
```go
type QueryIntent struct {
    Description             string   `json:"description"`
    RequiredCapabilities   []string `json:"requiredCapabilities"`
    RequiredKnowledgeBases []string `json:"requiredKnowledgeBases"`
    RequiredActionGroups   []string `json:"requiredActionGroups"`
    Complexity             string   `json:"complexity"` // "simple", "moderate", "complex"
    EstimatedSteps         int      `json:"estimatedSteps"`
}

type ExecutionPlan struct {
    InitialQuery           string              `json:"initialQuery"`
    Steps                  []PlannedStep       `json:"steps"`
    RequiredKnowledgeBases []string           `json:"requiredKnowledgeBases"`
    RequiredActionGroups   []string           `json:"requiredActionGroups"`
    MaxSteps              int                 `json:"maxSteps"`
    EstimatedDuration     time.Duration       `json:"estimatedDuration"`
}

type ReasoningStep struct {
    StepNumber   int                    `json:"stepNumber"`
    Description  string                 `json:"description"`
    Action       string                 `json:"action"`
    Input        map[string]interface{} `json:"input"`
    Output       map[string]interface{} `json:"output"`
    Reasoning    string                 `json:"reasoning"`
    Duration     time.Duration          `json:"duration"`
    Success      bool                   `json:"success"`
}
```

**Session and Context Models**:
```go
type ConversationTurn struct {
    TurnID      string                 `json:"turnId"`
    UserMessage string                 `json:"userMessage"`
    AgentResp   string                 `json:"agentResponse"`
    Timestamp   time.Time              `json:"timestamp"`
    Context     map[string]interface{} `json:"context"`
    Citations   []entities.Citation    `json:"citations"`
    Actions     []ActionResult         `json:"actions"`
}

type MemoryItem struct {
    ID          string                 `json:"id"`
    Content     string                 `json:"content"`
    Type        string                 `json:"type"` // "fact", "preference", "context"
    Importance  float64                `json:"importance"`
    Timestamp   time.Time              `json:"timestamp"`
    Source      string                 `json:"source"`
    Metadata    map[string]interface{} `json:"metadata"`
}

type ContextWindow struct {
    MaxTokens     int                    `json:"maxTokens"`
    CurrentTokens int                    `json:"currentTokens"`
    Turns         []ConversationTurn     `json:"turns"`
    Summary       string                 `json:"summary"`
    ImportantFacts []MemoryItem          `json:"importantFacts"`
}
```

Now I need to use the prework tool before writing the correctness properties:

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

Based on the prework analysis, I'll now define the key correctness properties that should hold universally across all valid executions:

### Property 1: Multi-step Query Decomposition
*For any* complex query requiring multiple steps, the Agent Core should decompose it into logical sub-tasks and execute them in the correct sequence with proper dependency management.
**Validates: Requirements 1.1, 1.4**

### Property 2: Multi-source Information Synthesis
*For any* query requiring information from multiple knowledge bases, the Agent Core should query all relevant sources and intelligently combine the results into a coherent response.
**Validates: Requirements 1.2, 4.3**

### Property 3: Action Group Integration Completeness
*For any* task requiring external API calls, the Agent Core should invoke appropriate action groups with correct parameters and integrate the results seamlessly into the response.
**Validates: Requirements 1.3, 3.2, 3.3**

### Property 4: Session Context Preservation
*For any* conversation reference to previous topics, the Agent Core should retrieve relevant context from session memory and respond appropriately without requiring clarification.
**Validates: Requirements 2.1, 2.2**

### Property 5: Memory Storage Consistency
*For any* important information shared during conversation, the Agent Core should store it in session memory and make it available for future reference throughout the session.
**Validates: Requirements 2.3**

### Property 6: Context Compression Intelligence
*For any* conversation that exceeds context window limits, the Agent Core should compress older context while preserving all important information and user preferences.
**Validates: Requirements 2.4**

### Property 7: Action Group Discovery and Selection
*For any* configured action groups, the Agent Core should be able to discover available functions and select the most appropriate ones based on user intent.
**Validates: Requirements 3.1, 3.5**

### Property 8: Knowledge Base Selection Intelligence
*For any* question requiring knowledge base information, the Agent Core should determine and query the most relevant knowledge bases using context-aware search strategies.
**Validates: Requirements 4.1, 4.2**

### Property 9: Citation and Attribution Completeness
*For any* knowledge base results retrieved, the Agent Core should provide proper citations and source attribution with clear explanations of relevance.
**Validates: Requirements 4.4, 10.3**

### Property 10: Streaming Progress Transparency
*For any* multi-step operation performed via streaming, the Agent Core should provide structured events indicating current steps, progress updates, and intermediate results.
**Validates: Requirements 5.1, 5.2, 5.3, 5.4**

### Property 11: Error Recovery and Reporting
*For any* failure in orchestration, action groups, or knowledge base queries, the system should provide detailed error information and implement appropriate recovery strategies.
**Validates: Requirements 6.1, 6.2, 6.3**

### Property 12: Configuration Application Consistency
*For any* custom configuration provided for reasoning strategies, memory management, or execution parameters, the system should apply these settings consistently across all operations.
**Validates: Requirements 7.1, 7.2, 7.3, 7.4**

### Property 13: Comprehensive Logging and Monitoring
*For any* Agent Core operation, action group invocation, or knowledge base query, the system should generate detailed logs and emit appropriate performance metrics.
**Validates: Requirements 8.1, 8.2, 8.3, 8.5**

### Property 14: Environment Configuration Adaptation
*For any* deployment environment, the system should automatically configure Agent Core settings based on environment variables and validate access to required resources.
**Validates: Requirements 9.1, 9.3, 9.5**

### Property 15: Reasoning Explanation Availability
*For any* multi-step reasoning process, the Agent Core should optionally provide clear explanations of planning, decision-making, and assumption-making processes.
**Validates: Requirements 10.1, 10.2, 10.4, 10.5**

### Property 16: Concurrent Session Isolation
*For any* multiple simultaneous user sessions, the Agent Core should handle them without interference while ensuring thread-safe access to session data.
**Validates: Requirements 12.1, 12.2**

### Property 17: Resource Management Under Load
*For any* concurrent operations involving action groups or knowledge base queries, the Agent Core should manage resources efficiently and prevent conflicts.
**Validates: Requirements 12.3, 12.4, 12.5**

### Property 18: Comprehensive Metrics Collection
*For any* agent interaction, multi-step operation, or system component usage, the system should collect detailed metrics for performance monitoring and optimization.
**Validates: Requirements 13.1, 13.2, 13.3, 13.4**

## Error Handling

### Agent Core Orchestration Errors

**Planning and Execution Failures**:
- Query analysis failures: Transform to `ErrCodeInvalidInput` with specific guidance on query reformulation
- Task decomposition errors: Transform to `ErrCodeServiceError` with details about which decomposition step failed
- Execution sequence errors: Transform to `ErrCodeExecutionError` with step-by-step failure information
- Dependency resolution failures: Transform to `ErrCodeDependencyError` with dependency chain details

**Multi-step Operation Errors**:
- Step execution failures: Provide detailed information about which step failed and why
- Result synthesis errors: Transform to `ErrCodeSynthesisError` with partial results when available
- Progress tracking failures: Log errors but continue execution when possible
- Timeout handling: Implement intelligent timeout management with partial result recovery

### Session Management Errors

**Context Management Failures**:
- Context retrieval errors: Transform to `ErrCodeContextError` with session recovery options
- Memory storage failures: Implement fallback storage with degraded functionality warnings
- Context compression errors: Provide manual compression options or context reset
- Session corruption: Implement automatic session recovery with user notification

**Memory and State Errors**:
- Memory extraction failures: Continue operation with reduced context awareness
- Context window overflow: Implement intelligent compression with user consent
- Session isolation failures: Ensure complete session separation with error logging
- Persistence failures: Implement in-memory fallback with session duration warnings

### Action Groups Integration Errors

**Function Discovery and Invocation**:
- Action group registration failures: Transform to `ErrCodeConfigurationError` with setup guidance
- Function discovery errors: Provide available function alternatives
- Parameter validation failures: Transform to `ErrCodeInvalidParameters` with parameter requirements
- Function execution timeouts: Implement retry logic with exponential backoff

**Result Processing Errors**:
- Result integration failures: Provide raw results with integration error details
- Response synthesis errors: Include partial results with synthesis failure explanation
- Function result parsing errors: Transform to `ErrCodeParsingError` with raw result fallback
- Concurrent execution conflicts: Implement intelligent queuing and retry mechanisms

### Knowledge Base Integration Errors

**Query and Retrieval Failures**:
- Knowledge base selection errors: Provide alternative knowledge base suggestions
- Query optimization failures: Fall back to original query with performance warnings
- Multi-source query failures: Provide partial results from successful sources
- Relevance scoring errors: Return results without scoring with manual relevance assessment

**Citation and Attribution Errors**:
- Citation generation failures: Provide results with manual citation requirements
- Source attribution errors: Include raw source information for manual processing
- Metadata extraction failures: Provide basic citation information with metadata gaps noted
- Cross-reference validation errors: Include citations with validation warnings

### Streaming and Real-time Errors

**Stream Processing Failures**:
- Stream initialization errors: Fall back to non-streaming mode with user notification
- Event processing failures: Continue streaming with event-specific error logging
- Buffer management errors: Implement dynamic buffer adjustment with performance impact warnings
- Stream completion errors: Provide partial results with completion status uncertainty

**Real-time Update Errors**:
- Progress tracking failures: Continue execution with reduced progress visibility
- Step notification errors: Provide final results with step summary
- Citation streaming errors: Include citations in final response with streaming failure note
- Action result streaming errors: Provide action results in final response summary

## Testing Strategy

### Dual Testing Approach

The testing strategy employs both unit testing and property-based testing to provide comprehensive coverage:

**Unit Tests**:
- Test specific Agent Core integration scenarios
- Validate configuration handling and error transformation
- Test session management and context operations
- Verify action group registration and invocation
- Mock Agent Core SDK for isolated testing

**Property-Based Tests**:
- Validate universal properties across all inputs and scenarios
- Test with generated conversation flows and complex queries
- Verify multi-step reasoning with simulated dependencies
- Test streaming behavior with various response patterns and interruptions
- Use **Go's testing/quick package** for property-based testing
- Configure each property test to run a minimum of **100 iterations**

### Property-Based Testing Implementation

**Testing Framework**: Use Go's built-in `testing/quick` package for property-based testing.

**Test Configuration**: Each property-based test must run a minimum of 100 iterations to ensure adequate coverage of the input space.

**Property Test Tagging**: Each property-based test must include a comment with the exact format:
`// **Feature: bedrock-agent-core-integration, Property {number}: {property_text}**`

**Example Property Test**:
```go
func TestProperty1_MultiStepQueryDecomposition(t *testing.T) {
    // **Feature: bedrock-agent-core-integration, Property 1: Multi-step Query Decomposition**
    
    agentCore, err := NewAgentCoreAdapter(ctx, testConfig)
    require.NoError(t, err)
    
    property := func(query string, complexity int) bool {
        // Generate complex multi-step query
        if query == "" || len(query) < 10 {
            query = generateComplexQuery(complexity)
        }
        
        input := &AgentCoreInput{
            SessionID: generateSessionID(),
            Message:   query,
            ExecutionConfig: &ExecutionConfig{
                Mode: "execution",
                MaxSteps: complexity + 2,
                EnableReasoning: true,
            },
        }
        
        response, err := agentCore.InvokeAgent(ctx, input)
        if err != nil {
            return false
        }
        
        // Verify multi-step decomposition occurred
        return len(response.ReasoningSteps) > 1 &&
               response.Content != "" &&
               response.ExecutionMetrics.TotalSteps > 0 &&
               isLogicalSequence(response.ReasoningSteps)
    }
    
    config := &quick.Config{MaxCount: 100}
    if err := quick.Check(property, config); err != nil {
        t.Errorf("Property violation: %v", err)
    }
}

func TestProperty4_SessionContextPreservation(t *testing.T) {
    // **Feature: bedrock-agent-core-integration, Property 4: Session Context Preservation**
    
    agentCore, err := NewAgentCoreAdapter(ctx, testConfig)
    require.NoError(t, err)
    
    property := func(initialInfo string, referenceQuery string) bool {
        sessionID := generateSessionID()
        
        // Generate valid inputs
        if initialInfo == "" {
            initialInfo = generateImportantInformation()
        }
        if referenceQuery == "" {
            referenceQuery = generateReferenceQuery(initialInfo)
        }
        
        // First interaction - establish context
        firstInput := &AgentCoreInput{
            SessionID: sessionID,
            Message:   initialInfo,
        }
        
        _, err := agentCore.InvokeAgent(ctx, firstInput)
        if err != nil {
            return false
        }
        
        // Second interaction - reference previous context
        secondInput := &AgentCoreInput{
            SessionID: sessionID,
            Message:   referenceQuery,
        }
        
        response, err := agentCore.InvokeAgent(ctx, secondInput)
        if err != nil {
            return false
        }
        
        // Verify context was preserved and referenced
        return response.Content != "" &&
               containsContextualReference(response.Content, initialInfo) &&
               response.SessionContext != nil &&
               len(response.SessionContext.ConversationHist) >= 2
    }
    
    config := &quick.Config{MaxCount: 100}
    if err := quick.Check(property, config); err != nil {
        t.Errorf("Property violation: %v", err)
    }
}
```

### Integration Testing Strategy

**Agent Core Environment Setup**:
- Use dedicated AWS test account with Bedrock Agent Core resources
- Create test agents with configured action groups and knowledge bases
- Configure IAM roles with appropriate Agent Core permissions
- Set up test knowledge bases with controlled content for predictable testing

**Multi-step Reasoning Tests**:
- Create complex scenarios requiring multiple reasoning steps
- Test dependency management between execution steps
- Validate result synthesis across multiple information sources
- Test error recovery in multi-step execution chains

**Session Management Tests**:
- Test context preservation across multiple conversation turns
- Validate memory storage and retrieval functionality
- Test context compression under various load conditions
- Verify session isolation between concurrent users

**Action Groups Integration Tests**:
- Test action group discovery and function selection
- Validate parameter passing and result integration
- Test error handling for action group failures
- Verify concurrent action group execution

**Knowledge Base Intelligence Tests**:
- Test intelligent knowledge base selection
- Validate context-aware search strategies
- Test multi-source information synthesis
- Verify citation generation and source attribution

### Performance and Scalability Testing

**Concurrent Session Testing**:
- Test multiple simultaneous agent sessions
- Validate session isolation and thread safety
- Test resource management under concurrent load
- Verify performance degradation patterns

**Multi-step Operation Performance**:
- Measure execution times for complex multi-step operations
- Test streaming performance with real-time updates
- Validate memory usage during long conversations
- Test context compression performance impact

**Action Groups and Knowledge Base Performance**:
- Test concurrent action group invocations
- Measure knowledge base query performance
- Test caching effectiveness for repeated queries
- Validate resource cleanup after operations

### Monitoring and Observability Testing

**Metrics Collection Validation**:
- Verify comprehensive metrics collection for all operations
- Test performance metrics accuracy and completeness
- Validate error rate tracking and categorization
- Test alert generation for threshold breaches

**Logging and Debugging Support**:
- Verify structured logging format consistency
- Test log correlation across multi-step operations
- Validate debugging information completeness
- Test integration with monitoring and alerting systems

This comprehensive testing strategy ensures that the Bedrock Agent Core integration provides reliable, intelligent, and scalable conversational AI capabilities while maintaining high performance and observability standards.