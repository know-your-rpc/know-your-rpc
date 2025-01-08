
const AWS = require('@aws-sdk/client-s3');
const fs = require('fs');

// @ts-ignore
const s3 = new AWS.S3(
    {
        region: 'eu-north-1',
    }
);

const uploadFile = async (filePath) => {
    const fileContent = fs.readFileSync(filePath);
    const params = {
        Bucket: "know-your-rpc-users-2",
        Key: "public.json",
        Body: fileContent
    };

    const command = new AWS.PutObjectCommand(params);

    await s3.send(command);

    console.log('success');
};

uploadFile('public.json');
