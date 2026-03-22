import client from './client';

// 意图分析
export interface IntentAnalysis {
  taskType: 'test' | 'code' | 'debug' | 'refactor' | 'document' | 'other';
  complexity: 'simple' | 'medium' | 'complex';
  scope: 'file' | 'package' | 'project';
  technologies: string[];
  dependencies: string[];
  constraints: string[];
  successCriteria: string[];
  rawQuery: string;
}

// 子任务
export interface SubTask {
  id: string;
  name: string;
  description: string;
  type: 'analyze' | 'create' | 'modify' | 'delete' | 'test' | 'verify';
  agentType: 'chat' | 'tool' | 'custom';
  tools: string[];
  dependsOn: string[];
  parallel: boolean;
}

// 工作流阶段
export interface WorkflowStage {
  type: 'sequential' | 'parallel' | 'task';
  task?: SubTask;
  tasks?: SubTask[];
  subStages?: WorkflowStage[];
}

// 工作流
export interface Workflow {
  type: 'sequential' | 'parallel' | 'task';
  stages: WorkflowStage[];
}

// 规划结果
export interface PlanResult {
  intent: IntentAnalysis;
  tasks: SubTask[];
  workflow: Workflow;
}

// 请求类型
export interface AnalyzeIntentRequest {
  query: string;
}

export interface AnalyzeIntentResponse {
  intent: IntentAnalysis;
}

export interface DecomposeTasksRequest {
  query: string;
  intent: IntentAnalysis;
}

export interface DecomposeTasksResponse {
  tasks: SubTask[];
}

export interface BuildWorkflowRequest {
  tasks: SubTask[];
}

export interface BuildWorkflowResponse {
  workflow: Workflow;
}

export interface PlanRequest {
  query: string;
}

export interface PlanResponse {
  intent: IntentAnalysis;
  tasks: SubTask[];
  workflow: Workflow;
}

// API 方法
export const plannerApi = {
  // 分析意图
  analyze: async (query: string): Promise<AnalyzeIntentResponse> => {
    return client.post('/planner/analyze', { query });
  },

  // 分解任务
  decompose: async (query: string, intent: IntentAnalysis): Promise<DecomposeTasksResponse> => {
    return client.post('/planner/decompose', { query, intent });
  },

  // 构建工作流
  buildWorkflow: async (tasks: SubTask[]): Promise<BuildWorkflowResponse> => {
    return client.post('/planner/workflow', { tasks });
  },

  // 完整规划
  plan: async (query: string): Promise<PlanResponse> => {
    return client.post('/planner/plan', { query });
  },
};

export default plannerApi;
