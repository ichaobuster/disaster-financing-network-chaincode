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

// const BlockDecoder = require('./node_modules/fabric-client/lib/BlockDecoder.js');
const BlockDecoder = require(path.join(__dirname, '/utils/blockDecoder.js'));
const pem = require('pem');
const cert = require('crypto').Certificate();


logger.info('First we enroll');
fcw.enrollWithAdminCert(helper.makeEnrollmentOptionsUsingCert(), function (enrollErr, enrollResp) {
	if (enrollErr != null) {
		logger.error('error enrolling', enrollErr, enrollResp);
	} else {
		const channel_id = helper.getChannelId();
		const first_peer = helper.getFirstPeerName(channel_id);
		const { client, channel } = enrollResp;

		const txId = client.newTransactionID();
		const request = {
			txId,
		};
		channel.getGenesisBlock(request).then(
			returnBlock => {
				const adminCerts = [];
				const blockData = BlockDecoder.decodeBlock(returnBlock);
				const groupData = blockData.data.data[0].payload.data.config.channel_group.groups[groupName].groups;
				const groupKeys = Object.keys(groupData);
				groupKeys.forEach(key => {
					adminCerts.push(groupData[key].values.MSP.value.config.admins[0]);
				});

				const orgInfos = [];
				adminCerts.forEach(rootCert => {
					pem.getPublicKey(rootCert, (error, keyInfo) => {
						const { publicKey } = keyInfo;
						pem.readCertificateInfo(rootCert, (error, certValues) => {
							const { commonName } = certValues;
							const organization = commonName.substring(commonName.lastIndexOf('@'));

							orgInfos.push({
								organization,
								publicKey,
							});

							console.info('orgInfos: ', orgInfos);
						})
					});
				});
			}
		);



	}
});
