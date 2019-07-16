//-------------------------------------------------------------------
// Chaincode Event
//-------------------------------------------------------------------
module.exports = function (logger) {
	var event_cc = {};

	event_cc.register_chaincode_event = function (obj, options, cb) {
		logger.debug('[fcw] Register Chaincode Event ' + options.event_name + ' for Chaincode: ' + options.chaincode_id);
		var channel = obj.channel;
		var client = obj.client;
		var eventHub;
		if (options.target_event_url && options.peer_tls_opts && options.event_name && options.chaincode_id) {
			logger.debug('[fcw] listening to chaincode event. url:', options.target_event_url);
			eventHub = client.newEventHub();
			eventHub.setPeerAddr(options.target_event_url, options.peer_tls_opts);
			eventHub.connect();

			// Wait for chaincode event - this will happen async
			eventHub.registerChaincodeEvent(options.chaincode_id, options.event_name, (chaincode_event) => {
				logger.info('Successfully got a chaincode event with transid:' + chaincode_event.tx_id);
				var event_payload = chaincode_event.payload.toString('utf8');
				if (cb) {
					cb(event_payload);
				}
			}, function (disconnectMsg) {											//callback whenever eventHub is disconnected, normal
				logger.info('[fcw] chaincode event ' + options.event_name + ' is disconnected');
			});
		} else {
			logger.debug('[fcw] will not use chaincode event');
		}
		return eventHub;
	};

	return event_cc;
};
