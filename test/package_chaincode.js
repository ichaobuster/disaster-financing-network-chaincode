const fs = require('fs');
const path = require('path');
const Golang = require('./node_modules/fabric-client/lib/packager/Golang.js');

const keep = [
	'.go',
	'.c',
	'.h'
];
const handler = new Golang(keep);

const outputPath = path.join(__dirname, '/chaincode.tar.gz');
process.env['GOPATH'] = path.join(__dirname, '..');;
const chaincodePath = 'dfn/go';
const metadataPath = '../src/dfn/go/META-INF';

handler.package(chaincodePath, metadataPath).then(
	data => {
		console.info('write file to: ', outputPath);
		fs.writeFileSync(outputPath, data);
	}
);
