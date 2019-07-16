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

console.log('---------------------------------------');
logger.info('Lets read -', chaincode_id, chaincode_ver);
console.log('---------------------------------------');
logger.warn('Note: the chaincode should have been installed before running this script');

logger.info('First we enroll');
fcw.enrollWithAdminCert(helper.makeEnrollmentOptionsUsingCert(), function (enrollErr, enrollResp) {
	if (enrollErr != null) {
		logger.error('error enrolling', enrollErr, enrollResp);
	} else {
		console.log('---------------------------------------');
		logger.info('Now we read');
		console.log('---------------------------------------');

		const channel = helper.getChannelId();
		const first_peer = helper.getFirstPeerName(channel);

		var id = 'project-bankcomm-000002';

		var opts = {
			peer_urls: [helper.getPeersUrl(first_peer)],
			peer_tls_opts: helper.getPeerTlsCertOpts(first_peer),
			channel_id: helper.getChannelId(),
			chaincode_id: chaincode_id,
			chaincode_version: chaincode_ver,
			cc_function: 'get_project_by_id',
			event_urls: ['grpc://localhost:7053'],
			cc_args: [
				id
			]
		};

		fcw.query_chaincode(enrollResp, opts, function (err, resp) {
			console.log('---------------------------------------');
			logger.info('read done. Errors:', (!err) ? 'nope' : err);
			console.log('---------------------------------------');
			if (!err && resp!=null && resp.peer_payloads!=null){
				logger.info('response is:', resp.peer_payloads[0]);
				console.log('---------------------------------------');
			}
		});

	}
});
