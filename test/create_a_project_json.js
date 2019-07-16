var winston = require('winston'); //logger module
var path = require('path');
var logger = new(winston.Logger)({
  level: 'debug',
  transports: [
    new(winston.transports.Console)({
      colorize: true
    }),
  ]
});

// --- Set Details Here --- //
var config_file = 'config_local.json'; //set config file name
var chaincode_id = 'test_abs_ledger'; //set desired chaincode id to identify this chaincode
var chaincode_ver = 'v0.0.1'; //set desired chaincode version

//  --- Use (optional) arguments if passed in --- //
var args = process.argv.slice(2);
if (args[0]) {
  config_file = args[0];
  logger.debug('Using argument for config file', config_file);
}
if (args[1]) {
  chaincode_id = args[1];
  logger.debug('Using argument for chaincode id');
}
if (args[2]) {
  chaincode_ver = args[2];
  logger.debug('Using argument for chaincode version');
}

var helper = require(path.join(__dirname, '/utils/helper.js'))(config_file, logger); //set the config file name here
var fcw = require(path.join(__dirname, '/utils/fc_wrangler/index.js'))({
  block_delay: helper.getBlockDelay()
}, logger);

console.log('---------------------------------------');
logger.info('Lets create a project -', chaincode_id, chaincode_ver);
console.log('---------------------------------------');
logger.warn('Note: the chaincode should have been installed before running this script');

logger.info('First we enroll');
fcw.enrollWithAdminCert(helper.makeEnrollmentOptionsUsingCert(), function(enrollErr, enrollResp) {
  if (enrollErr != null) {
    logger.error('error enrolling', enrollErr, enrollResp);
  } else {
    console.log('---------------------------------------');
    logger.info('Now we create a project');
    console.log('---------------------------------------');

    const channel = helper.getChannelId();
    const first_peer = helper.getFirstPeerName(channel);

    var projectInfo = {
      id: "project-bankcomm-000003",
      projectName: "测试交行项目000003号",
      scale: "10亿元人民币",
      basicAssets: "小微企业",
      initiator: "交通银行",
      trustee: "交银国信",
      depositary: "兴业银行",
      agent: "中债登",
      assetService: "上海融孚律师事务所",
      assessor: "深圳市世联资产评估有限公司",
      creditRater: "中债资信",
      liquiditySupporter: "中证信用增进股份有限公司",
      underwriter: "招商证券",
      lawyer: "北京市金杜律师事务所",
      accountant: "普华永道",
      createTime: new Date().toLocaleString("zh-CN")
    }

    var opts = {
      peer_urls: [helper.getPeersUrl(first_peer)],
      peer_tls_opts: helper.getPeerTlsCertOpts(first_peer),
      channel_id: helper.getChannelId(),
      chaincode_id: chaincode_id,
      chaincode_version: chaincode_ver,
      cc_function: 'create_project',
      event_urls: ['grpc://localhost:7053'],
      cc_args: [
        JSON.stringify(projectInfo)
      ]
    };

    fcw.invoke_chaincode(enrollResp, opts, function(err, resp) {
      console.log('---------------------------------------');
      logger.info('create a project done. Errors:', (!err) ? 'nope' : err);
      console.log('---------------------------------------');

      if (err != null) {
        return;
      }

      console.log('---------------------------------------');
      logger.info('Now we create another project');
      console.log('---------------------------------------');

      projectInfo.id = "project-bankcomm-000002";
      projectInfo.projectName = "测试交行项目000002号";
      projectInfo.scale = "500万元人民币";
      projectInfo.basicAssets = "个人按揭贷款";
      projectInfo.createTime = new Date().toLocaleString("zh-CN");

      opts = {
        peer_urls: [helper.getPeersUrl(first_peer)],
        peer_tls_opts: helper.getPeerTlsCertOpts(first_peer),
        channel_id: helper.getChannelId(),
        chaincode_id: chaincode_id,
        chaincode_version: chaincode_ver,
        cc_function: 'create_project',
        event_urls: ['grpc://localhost:7053'],
        cc_args: [
          JSON.stringify(projectInfo)
        ]
      };

      fcw.invoke_chaincode(enrollResp, opts, function(err, resp) {
        console.log('---------------------------------------');
        logger.info('create another project done. Errors:', (!err) ? 'nope' : err);
        console.log('---------------------------------------');
      });

    });

  }
});
