# Chaincode RSA API 文档

原则上，需要加密的数据、解密用的私钥不应该能被存储在CouchDB上和block上，所以这类数据需要使用``TransientMap``来传输。

## Invoke 可调用方法

### save_org_public_key

将公司数据加密公钥保存上链。

**参数：**
1. 公钥
2. 修改时间

**返回值：**
1. 无

**备注：**
1. 根据调用chaincode时发送的AdminCert中的CommonName解析和截取公司名称，形如：Admin@org1.example.com -> @org1.example.com

### encrypt_data

RSA加密存储数据上链（通用例子）。

**参数：**
1. ID（不可与链上任何已有ID重复，否则报错）
2. 需要加密数据的机构名清单，JSON字符串
3. 修改时间

**TransientMap：**
1. key为``dataString``
2. value为需要加密的数据，JSON字符串转换为bytes

**返回值：**
1. 无

**备注：**
1. 存储时，每一个机构存储一份``EncryptedData``数据。参见[EncryptedData的JSON说明](#encrypteddata的json说明)
2. 机构名清单形如 ```[@org1.example.com, @bankcomm.com]```
3. 默认加密时会加密一份chaincode操作提交机构的密文

### decrypt_data

RSA解密数据（通用例子）。

**参数：**
1. stateID

**TransientMap：**
1. key为``privateKey``
2. value为转换为bytes的私钥文件/字符串

**返回值：**
1. 解密后的JSON字符串

**备注：**
1. 无

## RSA加解密API

### Encrypt

RSA加密的底层API。

**参数：**
1. originData: 需要加密的原始数据，类型为[]byte
2. publicKey: 加密用的公钥，类型为[]byte

**返回值：**
1. 加密后的数据，类型为[]byte
2. 报错信息

### Encrypt64

RSA加密后对数据做base64编码。

**参数：**
1. originData: 需要加密的原始数据，类型为[]byte
2. publicKey: 加密用的公钥，类型为[]byte

**返回值：**
1. 加密后的数据编码为base64字符串，类型为string
2. 报错信息

### Decrypt

RSA解密的底层API。

**参数：**
1. cipherData: 加密数据，类型为[]byte
2. privateKey: 解密用的私钥，类型为[]byte

**返回值：**
1. 解密后的数据，类型为[]byte
2. 报错信息

### Decrypt64

RSA解密base64编码后的加密数据。

**参数：**
1. ciperData64: base64编码后的加密数据，类型为string
2. privateKey: 解密用的私钥，类型为[]byte

**返回值：**
1. 解密后的数据，类型为[]byte
2. 报错信息

### GetOrgPublicKeyId

生成机构公钥存储用的ID。

ID生成规则为：``org_public_key:`` ``机构标示`` 。例如：``org_public_key:@org1.example.com``

**参数：**
1. organization: 机构标示，如：@org1.example.com，类型为string

**返回值：**
1. 存储ID

### SaveOrgPublicKey

存储机构公钥。

如果机构公钥不存在，则直接存储公钥，版本设置为1.

如果机构公钥已存在，但公钥没有发生修改，则不做操作。

如果机构公钥已存在，且公钥发生修改，则存储新公钥后，升级版本号。

**参数：**
1. stub: shim.ChaincodeStubInterface
2. organization: 机构标示，如：@org1.example.com，类型为string
3. publicKey: 公钥字符串，类型为string
4. modifyTime: 修改时间字符串，类型为string

**返回值：**
1. 无

### GetOrgPublicKey

获取机构公钥信息。

**参数：**
1. stub: shim.ChaincodeStubInterface
2. organization: 机构标示，如：@org1.example.com，类型为string

**返回值：**
1. OrgPublicKey 结构体。参见[OrgPublicKey的JSON说明](#orgpublickey的json说明)

### GetStoreDataID

生成加密数据存储用的ID。

ID生成规则为：``stateID``:rsa:org:``机构标示`` 。例如：``xxxx-yyyy-zzzz-0000:rsa:org:@org1.example.com``

**参数：**
1. organization: 机构标示，如：@org1.example.com，类型为string

**返回值：**
1. 存储ID

### EncryptAndStore

使用公钥加密某机构数据，并存储到StateDB

**参数：**
1. stub: shim.ChaincodeStubInterface
2. stateID: 数据存储的原始ID，类型为string
3. originData: 需要加密的数据，类型为[]byte
4. organization: 机构标示，如：@org1.example.com，类型为string
5. modifyTime: 修改时间字符串，类型为string

**返回值：**
1. 错误信息

### EncryptAndStoreByOrgs

对列出的机构进行RSA公钥加密，并存储到StateDB

**参数：**
1. stub: shim.ChaincodeStubInterface
2. stateID: 数据存储的原始ID，类型为string
3. originData: 需要加密的数据，类型为[]byte
4. organizations: 机构标示列表，如：@org1.example.com，类型为[]string
5. modifyTime: 修改时间字符串，类型为string

**返回值：**
1. 错误信息

### GetEncryptData

获取加密数据结构体。

这里指的时获取当前操作机构的加密数据。

**参数：**
1. stub: shim.ChaincodeStubInterface
2. stateID: 数据存储的原始ID，类型为string

**返回值：**
1. ``EncryptedData``。参见[EncryptedData的JSON说明](#encrypteddata的json说明)
2. 错误信息

### GetEncryptTargetOrgs

获取对一个stateID进行过加密的机构清单。

**参数：**
1. stub: shim.ChaincodeStubInterface
2. stateID: 数据存储的原始ID，类型为string

**返回值：**
1. 机构标示列表，如：@org1.example.com，类型为[]string
2. 错误信息

## 其他

### EncryptedData的JSON说明

- **docType**: 资产类型，应为``encryptedData``
- **id**: 数据存储ID
- **originId**: 未加密的原始数据stateID
- **targetOrg**: 数据加密目标机构
- **encryptedData**: 加密数据（base64编码）
- **creator**: 创建人
- **lastModifier**: 最近修改人
- **createTime**: 创建时间
- **modifyTime**: 修改时间

### OrgPublicKey的JSON说明

- **docType**: 资产类型，应为``orgPublicKey``
- **id**: 公钥存储ID
- **organization**: 机构标示
- **publicKey**: 公钥字符串
- **version**: 版本号
- **createTime**: 创建时间
- **modifyTime**: 修改时间
