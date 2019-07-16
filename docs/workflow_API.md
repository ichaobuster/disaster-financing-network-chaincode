# Chaincode Workflow API 文档

本文档仅说明调用``Invoke``方法时可用的方法名和参数列表。

## create_linear_workflow

创建一个线性流程。

**参数：**
1. 描述工作流定义的JSON字符串。参见[workflowDef的JSON字段说明](#workflowdef的json字段说明)
2. 描述工作流节点的JSON字符串。参见[workflowNode的JSON字段说明](#workflownode的json字段说明)
3. ...

**返回值：**
1. 无

**备注：**

1. ``create_linear_workflow``接受任何大于等于2的参数个数
2. 前2个参数必须要有
3. 之后根据需要添加的节点数，重复添加第2个参数值即可

## get_workflow_by_id

使用ID查询一个工作流。

**参数：**
1. 工作流ID

**返回值：**
1. 描述一个流程的JSON，由``workflowDef``([workflowDef的JSON字段说明](#workflowdef的json字段说明))和``workflowNode``([workflowNode的JSON字段说明](#workflownode的json字段说明))组成

## query_all_workflows

~~**分页**~~ 查询所有工作流定义``workflowDef``。

**参数：**
1. ~~分页参数limit，表示每页多少条记录~~
2. ~~分页参数skip，表示调过前多少条记录~~

**返回值：**
1. 描述工作流定义``workflowDef``列表的JSON。参见[workflowDef的JSON字段说明](#workflowdef的json字段说明)

## query_accessable_workflows

~~**分页**~~ 查询所有可发起的工作流``workflowDef``。

**参数：**
1. ~~分页参数limit，表示每页多少条记录~~
2. ~~分页参数skip，表示调过前多少条记录~~

**返回值：**
1. 描述工作流定义``workflowDef``列表的JSON。参见[workflowDef的JSON字段说明](#workflowdef的json字段说明)


## enable_or_disable_workflow

停用或启用工作流。

**参数：**
1. 工作流ID
2. 启用或停用，使用 **字符串** "true"或"false"来控制

**返回值：**
无

**备注：**

1. 目前限定只有创建人所在机构能够启停用工作流

## modify_workflow_def

修改一个工作流信息``workflowDef``。

**参数:**
1. 工作流ID
2. 前4个参数必须要有
3. 要修改的字段名，参见[workflowDef的JSON字段说明](#workflowdef的json字段说明)
4. 修改后的值
5. ...

**返回值：**
1. 无

**备注：**

1. ``modify_workflow_def``接受任何大于等于3的 **偶数** 参数个数
2. 前4个参数必须要有
3. 之后根据需要修改的字段数，重复添加第3和第4个参数值即可
4. 用例可参照[``modify_project``](project_API.md#modify_project)

## 其他

### workflowDef的JSON字段说明

- **docType**: 资产类型，应为``workflow``，不可修改该字段值
- **id**: 工作流ID，不可修改该字段值
- **subDocType**: 资产副类型，线性流程为``linearWorkflow``，不可修改该字段值
- **workflowName**: 工作流名称
- **accessRoles**: 字符串数组，指定可发起流程的角色
- **accessOrgs**: 字符串数组，指定可发起流程的机构
- **enabled**: bool类型，是否启用工作流，不可使用修改方法来修改该字段值
- **creator**: 创建人，不可修改该字段值
- **lastModifier**: 最近修改人，不可修改该字段值
- **createTime**: 创建时间
- **modifyTime**: 修改时间

### workflowNode的JSON字段说明

- **docType**: 资产类型，应为``workflowNode``，不可修改该字段值
- **id**: 工作流节点ID，自动生成，不可修改该字段值
- **workflowId**: 工作流ID
- **nodeName**: 工作流节点名称
- **accessRoles**: 字符串数组，指定可使用该节点的角色
- **accessOrgs**: 字符串数组，指定可使用该节点的机构
- **PrevNodeIds**: 上一节点ID；对于线性流程，该字段自动生成
- **nextNodeIds**: 下一节点ID；对于线性流程，该字段自动生成
- **firstNode**: bool型，是否是第一个节点；对于线性流程，该字段自动生成
- **lastNode**: bool型，是否是最后一个节点；对于线性流程，该字段自动生成
