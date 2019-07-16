const blockDecoder = require('../node_modules/fabric-client/lib/BlockDecoder.js');

module.exports = {
  decode: (blockBytes) => {
    return blockDecoder.decode(blockBytes);
  },

  decodeBlock: (blockData) => {
    return blockDecoder.decodeBlock(blockData);
  },

  decodeTransaction: (processedTransactionBytes) => {
    return blockDecoder.decodeTransaction(processedTransactionBytes);
  },
}
