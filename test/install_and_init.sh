#!/bin/bash

# 安装chaincode
node install_chaincode.js
# chaincode实例化
node instantiate_chaincode.js
# 保存公钥
node save_rsa_pubkey.js
# 创建项目（2个）
node create_a_project_json.js
# 创建一个线性流程
node create_a_linear_workflow.js
# 创建流程实例（2个）
node start_a_process.js
# 流转流程
node transfer_process.js
# 将流程流转至完成
node transfer_process_to_finish.js
# 退回流程
# node return_process.js
