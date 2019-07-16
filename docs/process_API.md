# Chaincode Process API 文档

本文档仅说明调用``Invoke``方法时可用的方法名和参数列表。

## start_process

开始一个流程。

**参数：**
1. 描述流程开始信息的JSON字符串。参见[process的JSON字段说明](#process的json字段说明)

**返回值：**
1. 无

**备注：**

1. 必要字段：
  - id
  - workflowId
  - attachDocType
  - attachDocId

## get_process_by_id

获取一个流程的详细信息。

**参数：**
1. 流程实例ID

**返回值：**
1. 描述一个流程实例的JSON。参见[process的JSON字段说明](#process的json字段说明)

## query_logs_by_process_id

获取一个流程的流转日志。

**参数：**
1. 流程实例ID

**返回值：**
1. 描述流转日志的JSON数组。参见[processLog的JSON字段说明](#processlog的json字段说明)

## transfer_process

流程实例运行/传递。

**参数：**
1. 流程实例ID
2. 下一节点ID。如果当前节点是最后一个节点（LastNode == true），此项参数无效。
3. 下一拥有人/机构。如果当前节点是最后一个节点（LastNode == true），此项参数无效。

**返回值：**
1. 无

**备注：**

1. 如果当前节点是最后一个节点（LastNode == true），下一节点ID和下一拥有人/机构参数无效，可填写任意值。
2. 如果当前节点是最后一个节点（LastNode == true），流程将直接提交至办结。

## cancel_process

取消流程实例。

**参数：**
1. 流程实例ID

**返回值：**
1. 无

**备注：**

1. 已取消、已完成的流程不能取消
2. 目前限定只有流程实例创建人所在机构能取消流程实例

## query_todo_process

~~**分页**~~ 查询待办流程实例。

**参数：**
1. ~~分页参数limit，表示每页多少条记录~~
2. ~~分页参数skip，表示调过前多少条记录~~

**返回值：**
1. 描述流程实例列表的JSON。参见[process的JSON字段说明](#process的json字段说明)

## query_done_process

~~**分页**~~ 查询已办流程实例。

**参数：**
1. ~~分页参数limit，表示每页多少条记录~~
2. ~~分页参数skip，表示调过前多少条记录~~

**返回值：**
1. 描述流程实例列表的JSON。参见[process的JSON字段说明](#process的json字段说明)

## 其他

### process的JSON字段说明

- **docType**: 资产类型，应为``process``
- **id**: 流程实例ID
- **attachDocType**: 附加在流程上的文档类型
- **attachDocId**: 附加在流程上的文档ID
- **attachDocName**: 附加在流程上的文档名称
- **workflowId**: 工作流ID
- **workflowName**: 工作流名称
- **currentNodeId**: 当前节点ID
- **currentNodeName**: 当前节点名称
- **currentOwner**: 当前拥有人/机构
- **participants**: 已参与流程流转的参与人清单
- **finished**: bool型，是否已完成
- **canceled**: bool型，是否已取消
- **creator**: 流程创建人
- **lastModifier**: 最近修改人
- **createTime**: 创建时间
- **modifyTime**: 修改时间

### processLog的JSON字段说明
- **docType**: 资产类型，应为``processLog``
- **id**: 日志记录ID
- **processId**: 流程实例ID
- **fromNodeId**: 提交方节点ID
- **fromNodeName**: 提交放节点名称
- **fromOrg**: 提交方机构
- **toNodeId**: 接收方节点ID
- **toNodeName**: 接收方节点名称
- **toOrg**: 接收方机构
- **remark**: 备注
- **createTime**: 创建时间
- **modifyTime**: 修改时间
