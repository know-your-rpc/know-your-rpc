
const AWS = require('@aws-sdk/client-s3');
const fs = require('fs');
const path = require('path');


// @ts-ignore
const s3 = new AWS.S3(
    // if it works remove comment
    //   {
    // region: 'eu-north-1',
    // credentials: {
    //     accessKeyId: process.env['AWS_ACCESS_KEY_ID'],
    //     secretAccessKey: process.env['AWS_SECRET_ACCESS_KEY']
    // }
    //}
);

const uploadFile = async (filePath) => {
    const fileContent = fs.readFileSync(filePath);
    const params = {
        Bucket: "know-your-rpc-users",
        Key: "public.json",
        Body: fileContent
    };

    const command = new AWS.PutObjectCommand(params);

    await s3.send(command);

    console.log('success');
};

uploadFile('public.json');
