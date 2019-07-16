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
var fs = require('fs');

var path2cc = path.join(__dirname, '/chaincode.tar.gz');

console.log('---------------------------------------');
logger.info('Lets install some chaincode -', chaincode_id, chaincode_ver);
console.log('---------------------------------------');

logger.info('First we enroll');
fcw.enrollWithAdminCert(helper.makeEnrollmentOptionsUsingCert(), function (enrollErr, enrollResp) {
	if (enrollErr != null) {
		logger.error('error enrolling', enrollErr, enrollResp);
	} else {
		console.log('---------------------------------------');
		logger.info('Now we install');
		console.log('---------------------------------------');

		const channel = helper.getChannelId();
		const first_peer = helper.getFirstPeerName(channel);

		const cc_bytes = fs.readFileSync(path2cc);

		var opts = {
			peer_urls: [helper.getPeersUrl(first_peer)],
			path_2_chaincode: 'dfn/go',
			chaincode_id: chaincode_id,
			chaincode_version: chaincode_ver,
			chaincode_type: 'golang',
			peer_tls_opts: helper.getPeerTlsCertOpts(first_peer),
			chaincode_package: cc_bytes,
		};
		fcw.install_chaincode(enrollResp, opts, function (err, resp) {
			console.log('---------------------------------------');
			logger.info('Install done. Errors:', (!err) ? 'nope' : err);
			console.log('---------------------------------------');
		});
	}
});
