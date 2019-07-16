# Chaincode Project API 文档

本文档仅说明调用``Invoke``方法时可用的方法名和参数列表。

## create_project

使用JSON创建一个项目。

**参数：**
1. 描述项目的JSON字符串。参见[project的JSON字段说明](#project的json字段说明)

**返回值：**
1. 无

## remove_project

删除一个项目。

**参数：**
1. 项目ID

**返回值：**
1. 无

## get_project_by_id

使用ID查询一个项目``project``。

**参数：**
1. 项目ID

**返回值：**
1. 描述一个项目``project``的JSON。参见[project的JSON字段说明](#project的json字段说明)

## query_all_projects

~~**分页**~~ 查询所有项目``project``。

**参数：**
1. ~~分页参数limit，表示每页多少条记录~~
2. ~~分页参数skip，表示调过前多少条记录~~

**返回值：**
1. 描述项目``project``列表的JSON。参见[project的JSON字段说明](#project的json字段说明)

## query_all_projects

**分页** 查询所有项目``project``。

**参数：**
1. 分页参数limit，表示每页多少条记录
2. 分页参数skip，表示调过前多少条记录

**返回值：**
1. 描述项目``project``列表的JSON。参见[project的JSON字段说明](#project的json字段说明)

## modify_project

修改一个项目``project``。

**参数:**
1. 项目ID
2. 修改时间
3. 要修改的字段名，参见[project的JSON字段说明](#project的json字段说明)
4. 修改后的值
5. ...

**返回值：**
1. 无

**备注：**

1. ``modify_project``接受任何大于等于3的 **偶数** 参数个数
2. 前4个参数必须要有
3. 之后根据需要修改的字段数，重复添加第3和第4个参数值即可
4. 目前定义为仅有创建人可修改

例如，同时修改项目``test-project-000001``的``项目名称``和``发行概况``，参数值可以定义为：

````
[
  "test-project-000001",
  "2018-3-3 10:00:00"
  "projectName",
  "新项目名",
  "scale",
  "10亿元人民币"
]
````

## 其他

### project的JSON字段说明

- **docType**: 资产类型，应为``project``，不可修改该字段值
- **id**: 项目ID，不可修改该字段值
- **projectName**: 项目名称
- **scale**: 发行规模
- **basicAssets**: 基础资产
- **overview**: 发行概况
- **initiator**: 发起机构
- **trustee**: 受托机构
- **depositary**: 资金保管机构
- **agent**: 登记/支付代理机构
- **assetService**: 资产服务机构
- **assessor**: 评估机构
- **creditRater**: 信用评级机构
- **liquiditySupporter**: 流动性支持机构
- **underwriter**: 承销商/薄记管理人机构
- **lawyer**: 律师
- **accountant**: 会计师
- **creator**: 创建人，不可修改该字段值
- **lastModifier**: 最近修改人，不可修改该字段值
- **createTime**: 创建时间
- **modifyTime**: 修改时间
