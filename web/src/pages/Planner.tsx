import React, { useState } from 'react';
import {
  Card,
  Typography,
  Input,
  Button,
  Space,
  Tabs,
  Tag,
  List,
  Steps,
  Divider,
  Alert,
  Descriptions,
  Grid,
  Spin,
  message,
  Collapse,
} from 'antd';
import {
  ThunderboltOutlined,
  ApiOutlined,
  ExperimentOutlined,
  CodeOutlined,
  CheckCircleOutlined,
  PlayCircleOutlined,
  BulbOutlined,
  TeamOutlined,
  ShareAltOutlined,
} from '@ant-design/icons';
import plannerApi, {
  type IntentAnalysis,
  type SubTask,
  type Workflow,
  type PlanResult,
} from '../api/planner';

const { Title, Text, Paragraph } = Typography;
const { TextArea } = Input;
const { TabPane } = Tabs;
const { useBreakpoint } = Grid;
const { Panel } = Collapse;

// 示例查询
const EXAMPLE_QUERIES = [
  '为用户服务创建单元测试',
  '创建一个新的 API 接口实现',
  '重构对话记录模块代码',
  '修复登录功能的 bug',
  '为项目编写 README 文档',
];

const Planner: React.FC = () => {
  const screens = useBreakpoint();
  const [query, setQuery] = useState('');
  const [loading, setLoading] = useState(false);
  const [step, setStep] = useState(0);
  const [intentResult, setIntentResult] = useState<IntentAnalysis | null>(null);
  const [tasksResult, setTasksResult] = useState<SubTask[] | null>(null);
  const [workflowResult, setWorkflowResult] = useState<Workflow | null>(null);
  const [planResult, setPlanResult] = useState<PlanResult | null>(null);
  const [activeTab, setActiveTab] = useState('full');

  // 任务类型标签颜色
  const taskTypeColors: Record<string, string> = {
    analyze: 'blue',
    create: 'green',
    modify: 'orange',
    delete: 'red',
    test: 'cyan',
    verify: 'purple',
  };

  // 任务类型标签文本
  const taskTypeLabels: Record<string, string> = {
    analyze: '分析',
    create: '创建',
    modify: '修改',
    delete: '删除',
    test: '测试',
    verify: '验证',
  };

  // Agent 类型标签
  const agentTypeLabels: Record<string, string> = {
    chat: '对话',
    tool: '工具',
    custom: '自定义',
  };

  // 复杂度标签
  const complexityLabels: Record<string, string> = {
    simple: '简单',
    medium: '中等',
    complex: '复杂',
  };

  // 范围标签
  const scopeLabels: Record<string, string> = {
    file: '文件',
    package: '包',
    project: '项目',
  };

  // 任务类型图标
  const getTaskTypeIcon = (type: string) => {
    switch (type) {
      case 'analyze':
        return <BulbOutlined />;
      case 'create':
        return <CodeOutlined />;
      case 'modify':
        return <ApiOutlined />;
      case 'test':
        return <ExperimentOutlined />;
      case 'verify':
        return <CheckCircleOutlined />;
      default:
        return <ThunderboltOutlined />;
    }
  };

  // 分析意图
  const handleAnalyzeIntent = async () => {
    if (!query.trim()) {
      message.warning('请输入查询内容');
      return;
    }
    setLoading(true);
    setStep(0);
    try {
      const res = await plannerApi.analyze(query);
      setIntentResult(res.intent);
      setStep(1);
      message.success('意图分析完成');
    } catch (error) {
      console.error('意图分析失败:', error);
      message.error('意图分析失败');
    } finally {
      setLoading(false);
    }
  };

  // 分解任务
  const handleDecomposeTasks = async () => {
    if (!intentResult) return;
    setLoading(true);
    try {
      const res = await plannerApi.decompose(query, intentResult);
      setTasksResult(res.tasks);
      setStep(2);
      message.success('任务分解完成');
    } catch (error) {
      console.error('任务分解失败:', error);
      message.error('任务分解失败');
    } finally {
      setLoading(false);
    }
  };

  // 构建工作流
  const handleBuildWorkflow = async () => {
    if (!tasksResult) return;
    setLoading(true);
    try {
      const res = await plannerApi.buildWorkflow(tasksResult);
      setWorkflowResult(res.workflow);
      setStep(3);
      message.success('工作流构建完成');
    } catch (error) {
      console.error('工作流构建失败:', error);
      message.error('工作流构建失败');
    } finally {
      setLoading(false);
    }
  };

  // 完整规划
  const handleFullPlan = async () => {
    if (!query.trim()) {
      message.warning('请输入查询内容');
      return;
    }
    setLoading(true);
    setStep(0);
    try {
      const res = await plannerApi.plan(query);
      setPlanResult(res);
      setIntentResult(res.intent);
      setTasksResult(res.tasks);
      setWorkflowResult(res.workflow);
      setStep(3);
      message.success('完整规划完成');
    } catch (error) {
      console.error('完整规划失败:', error);
      message.error('完整规划失败');
    } finally {
      setLoading(false);
    }
  };

  // 重置
  const handleReset = () => {
    setIntentResult(null);
    setTasksResult(null);
    setWorkflowResult(null);
    setPlanResult(null);
    setStep(0);
  };

  // 使用示例
  const useExample = (example: string) => {
    setQuery(example);
    handleReset();
  };

  // 渲染意图分析结果
  const renderIntentResult = () => {
    if (!intentResult) return null;
    return (
      <Card
        title={<Space><BulbOutlined /> 意图分析结果</Space>}
        style={{ marginTop: 16 }}
        size="small"
      >
        <Descriptions column={screens.xs ? 1 : 2} bordered size="small">
          <Descriptions.Item label="任务类型">
            <Tag color="blue">{intentResult.taskType}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="复杂度">
            <Tag color={intentResult.complexity === 'simple' ? 'green' : intentResult.complexity === 'medium' ? 'orange' : 'red'}>
              {complexityLabels[intentResult.complexity]}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="范围">
            <Tag color="purple">{scopeLabels[intentResult.scope]}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="技术栈">
            <Space wrap>
              {intentResult.technologies.map((tech, i) => (
                <Tag key={i} color="cyan">{tech}</Tag>
              ))}
            </Space>
          </Descriptions.Item>
          <Descriptions.Item label="依赖" span={screens.xs ? 1 : 2}>
            {intentResult.dependencies.length > 0 ? (
              <Space wrap>
                {intentResult.dependencies.map((dep, i) => (
                  <Tag key={i}>{dep}</Tag>
                ))}
              </Space>
            ) : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="约束" span={screens.xs ? 1 : 2}>
            {intentResult.constraints.length > 0 ? (
              <List
                size="small"
                dataSource={intentResult.constraints}
                renderItem={(item) => <List.Item>• {item}</List.Item>}
              />
            ) : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="成功标准" span={screens.xs ? 1 : 2}>
            <List
              size="small"
              dataSource={intentResult.successCriteria}
              renderItem={(item) => <List.Item><CheckCircleOutlined style={{ color: '#52c41a', marginRight: 8 }} />{item}</List.Item>}
            />
          </Descriptions.Item>
        </Descriptions>
      </Card>
    );
  };

  // 渲染任务列表
  const renderTasksResult = () => {
    if (!tasksResult || tasksResult.length === 0) return null;
    return (
      <Card
        title={<Space><TeamOutlined /> 分解任务 ({tasksResult.length} 个)</Space>}
        style={{ marginTop: 16 }}
        size="small"
      >
        <List
          grid={{ gutter: 16, xs: 1, sm: 1, md: 1, lg: 2, xl: 2, xxl: 2 }}
          dataSource={tasksResult}
          renderItem={(task) => (
            <List.Item>
              <Card
                size="small"
                type="inner"
                title={
                  <Space>
                    {getTaskTypeIcon(task.type)}
                    <Text strong>{task.name}</Text>
                    <Tag color={taskTypeColors[task.type]}>{taskTypeLabels[task.type]}</Tag>
                    <Tag color="default">{agentTypeLabels[task.agentType]} Agent</Tag>
                  </Space>
                }
              >
                <Paragraph ellipsis={{ rows: 2 }} style={{ marginBottom: 8 }}>
                  {task.description}
                </Paragraph>
                {task.tools && task.tools.length > 0 && (
                  <Space wrap style={{ marginTop: 8 }}>
                    <Text type="secondary" style={{ fontSize: '12px' }}>工具:</Text>
                    {task.tools.map((tool, i) => (
                      <Tag key={i} size="small">{tool}</Tag>
                    ))}
                  </Space>
                )}
                {task.dependsOn && task.dependsOn.length > 0 && (
                  <Space wrap style={{ marginTop: 8 }}>
                    <Text type="secondary" style={{ fontSize: '12px' }}>依赖:</Text>
                    {task.dependsOn.map((dep, i) => (
                      <Tag key={i} size="small" color="orange">{dep}</Tag>
                    ))}
                  </Space>
                )}
                {task.parallel && (
                  <Tag color="green" style={{ marginTop: 8 }}>可并行执行</Tag>
                )}
              </Card>
            </List.Item>
          )}
        />
      </Card>
    );
  };

  // 渲染工作流阶段
  const renderWorkflowStage = (stage: any, level: number = 0) => {
    const indent = level * 24;
    return (
      <div key={`stage-${Math.random()}`} style={{ marginLeft: indent }}>
        {stage.type === 'task' && stage.task && (
          <Steps.Step
            title={stage.task.name}
            description={
              <Space>
                <Tag size="small" color={taskTypeColors[stage.task.type]}>
                  {taskTypeLabels[stage.task.type]}
                </Tag>
                <Text type="secondary" ellipsis={{ rows: 1 }} style={{ fontSize: '12px' }}>
                  {stage.task.description}
                </Text>
              </Space>
            }
            icon={getTaskTypeIcon(stage.task.type)}
          />
        )}
        {stage.type === 'sequential' && (
          <>
            <div style={{ marginBottom: 8 }}>
              <Tag color="blue">顺序执行</Tag>
            </div>
            {stage.subStages?.map((sub: any) => renderWorkflowStage(sub, level + 1))}
          </>
        )}
        {stage.type === 'parallel' && (
          <>
            <div style={{ marginBottom: 8 }}>
              <Tag color="green">并行执行</Tag>
            </div>
            {stage.subStages?.map((sub: any) => renderWorkflowStage(sub, level + 1))}
          </>
        )}
        {stage.tasks && stage.tasks.length > 0 && (
          <>
            {stage.tasks.map((task: any) => (
              <Steps.Step
                key={task.id}
                title={task.name}
                description={
                  <Space>
                    <Tag size="small" color={taskTypeColors[task.type]}>
                      {taskTypeLabels[task.type]}
                    </Tag>
                    <Text type="secondary" ellipsis={{ rows: 1 }} style={{ fontSize: '12px' }}>
                      {task.description}
                    </Text>
                  </Space>
                }
                icon={getTaskTypeIcon(task.type)}
              />
            ))}
          </>
        )}
      </div>
    );
  };

  // 渲染工作流
  const renderWorkflowResult = () => {
    if (!workflowResult) return null;
    return (
      <Card
        title={<Space><ShareAltOutlined /> 工作流编排</Space>}
        style={{ marginTop: 16 }}
        size="small"
      >
        <Alert
          message={
            <Space>
              {workflowResult.type === 'sequential' ? '顺序模式' : '并行模式'}
            </Space>
          }
          type="info"
          showIcon
          style={{ marginBottom: 16 }}
        />
        <Steps direction="vertical" size="small">
          {workflowResult.stages?.map((stage, index) => renderWorkflowStage(stage, 0))}
        </Steps>
      </Card>
    );
  };

  // 分步执行模式
  const renderStepMode = () => (
    <div>
      <Card
        title="分步执行"
        extra={
          <Space>
            <Button onClick={handleReset}>重置</Button>
          </Space>
        }
      >
        <Steps current={step} style={{ marginBottom: 24 }}>
          <Steps.Step title="意图分析" icon={<BulbOutlined />} />
          <Steps.Step title="任务分解" icon={<TeamOutlined />} />
          <Steps.Step title="工作流编排" icon={<ShareAltOutlined />} />
        </Steps>

        <Space direction="vertical" style={{ width: '100%' }}>
          {step === 0 && (
            <Space>
              <Button
                type="primary"
                icon={<PlayCircleOutlined />}
                onClick={handleAnalyzeIntent}
                loading={loading}
              >
                分析意图
              </Button>
            </Space>
          )}
          {step === 1 && (
            <Space>
              <Button
                type="primary"
                icon={<PlayCircleOutlined />}
                onClick={handleDecomposeTasks}
                loading={loading}
              >
                分解任务
              </Button>
            </Space>
          )}
          {step === 2 && (
            <Space>
              <Button
                type="primary"
                icon={<PlayCircleOutlined />}
                onClick={handleBuildWorkflow}
                loading={loading}
              >
                构建工作流
              </Button>
            </Space>
          )}
          {step === 3 && (
            <Alert
              message="分步执行完成！"
              description="所有步骤已执行完毕，可以查看下方的结果展示。"
              type="success"
              showIcon
            />
          )}
        </Space>
      </Card>

      {renderIntentResult()}
      {renderTasksResult()}
      {renderWorkflowResult()}
    </div>
  );

  // 完整规划模式
  const renderFullMode = () => (
    <div>
      <Card
        title="完整规划"
        extra={
          <Space>
            <Button onClick={handleReset}>重置</Button>
            <Button
              type="primary"
              icon={<ThunderboltOutlined />}
              onClick={handleFullPlan}
              loading={loading}
            >
              一键规划
            </Button>
          </Space>
        }
      >
        {step === 3 && planResult ? (
          <Alert
            message="规划完成！"
            description="已成功完成意图分析、任务分解和工作流编排。"
            type="success"
            showIcon
          />
        ) : (
          <Alert
            message="点击「一键规划」执行完整的规划流程"
            description="系统将依次执行意图分析、任务分解和工作流编排三个步骤。"
            type="info"
            showIcon
          />
        )}
      </Card>

      {renderIntentResult()}
      {renderTasksResult()}
      {renderWorkflowResult()}
    </div>
  );

  return (
    <div>
      <Card
        title={
          <Space>
            <Title level={screens.xs ? 4 : 3} style={{ margin: 0 }}>
              Eino 自动规划系统
            </Title>
          </Space>
        }
        style={{ marginBottom: 16 }}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <Text type="secondary">
            输入您的任务描述，系统将自动分析意图、分解任务并编排工作流。
          </Text>
          <TextArea
            rows={4}
            placeholder="请输入任务描述，例如：为用户模块创建单元测试、创建一个新的 API 接口..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
          <Space wrap>
            <Text type="secondary" style={{ fontSize: '12px' }}>示例:</Text>
            {EXAMPLE_QUERIES.map((example) => (
              <Button
                key={example}
                size="small"
                type="link"
                onClick={() => useExample(example)}
              >
                {example}
              </Button>
            ))}
          </Space>
        </Space>
      </Card>

      <Card>
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={[
            {
              key: 'full',
              label: (
                <span>
                  <ThunderboltOutlined />
                  一键规划
                </span>
              ),
            },
            {
              key: 'step',
              label: (
                <span>
                  <PlayCircleOutlined />
                  分步执行
                </span>
              ),
            },
          ]}
        />
        <Divider style={{ margin: '16px 0' }} />
        {activeTab === 'full' ? renderFullMode() : renderStepMode()}
      </Card>
    </div>
  );
};

export default Planner;
