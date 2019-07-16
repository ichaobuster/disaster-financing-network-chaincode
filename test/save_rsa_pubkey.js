var winston = require('winston');								//logger module
var path = require('path');
var logger = new (winston.Logger)({
	level: 'debug',
	transports: [
		new (winston.transports.Console)({ colorize: true }),
	]
});

// --- Set Details Here --- //
var config_file = 'config_local.json';					//set config file name
var chaincode_id = 'test_abs_ledger';						//set desired chaincode id to identify this chaincode
var chaincode_ver = 'v0.0.1';										//set desired chaincode version
// 根据实际情况修改
const groupName = 'Application';
const genesisName = 'Org1MSP';

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

var helper = require(path.join(__dirname, '/utils/helper.js'))(config_file, logger);			//set the config file name here
var fcw = require(path.join(__dirname, '/utils/fc_wrangler/index.js'))({ block_delay: helper.getBlockDelay() }, logger);


const publicKey = '-----BEGIN PUBLIC KEY-----\nMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDZsfv1qscqYdy4vY+P4e3cAtmv\nppXQcRvrF1cB4drkv0haU24Y7m5qYtT52Kr539RdbKKdLAM6s20lWy7+5C0Dgacd\nwYWd/7PeCELyEipZJL07Vro7Ate8Bfjya+wltGK9+XNUIHiumUKULW4KDx21+1NL\nAUeJ6PeW+DAkmJWF6QIDAQAB\n-----END PUBLIC KEY-----';

logger.info('First we enroll');
fcw.enrollWithAdminCert(helper.makeEnrollmentOptionsUsingCert(), function (enrollErr, enrollResp) {
	if (enrollErr != null) {
		logger.error('error enrolling', enrollErr, enrollResp);
	} else {
		console.log('---------------------------------------');
		logger.info('Now we start save rsa pubkey');
		console.log('---------------------------------------');

		const channel = helper.getChannelId();
		const first_peer = helper.getFirstPeerName(channel);

		const modifyTime = new Date().toLocaleString("zh-CN");

		var opts = {
			peer_urls: [helper.getPeersUrl(first_peer)],
			peer_tls_opts: helper.getPeerTlsCertOpts(first_peer),
			channel_id: helper.getChannelId(),
			chaincode_id: chaincode_id,
			chaincode_version: chaincode_ver,
			cc_function: 'save_org_public_key',
			event_urls: ['grpc://localhost:7053'],
			cc_args: [
				publicKey,
				modifyTime
			],
		};

		fcw.invoke_chaincode(enrollResp, opts, function (err, resp) {
			console.log('---------------------------------------');
			logger.info('save rsa pubkey done. Errors:', (!err) ? 'nope' : err);
			console.log('---------------------------------------');
		});
	}
});
