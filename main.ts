import { Construct } from "constructs";
import { App, AssetType, TerraformAsset, TerraformStack } from "cdktf";
import * as google from '@cdktf/provider-google';
import * as path from 'path';

const project = 'ubiquitous-spoon';
const region = 'us-central1';
const repository = 'ubiquitous-spoon';

class MyStack extends TerraformStack {
  constructor(scope: Construct, id: string) {
    super(scope, id);

    new google.provider.GoogleProvider(this, 'google', {
      project,
    });

    const functionRunner = new google.serviceAccount.ServiceAccount(this, 'functionRunner', {
      accountId: 'function-runner',
    });

    const assetBucket = new google.storageBucket.StorageBucket(this, 'assetBucket', {
      lifecycleRule: [{
        action: {
          type: 'delete',
        },
        condition: {
          age: 1,
        },
      }],
      location: region,
      name: `asset-bucket-${project}`,
    });

    const sourceBucket = new google.storageBucket.StorageBucket(this, 'sourceBucket', {
      lifecycleRule: [{
        action: {
          type: 'delete',
        },
        condition: {
          age: 1,
        },
      }],
      location: region,
      name: `source-bucket-${project}`,
    });

    const destinationBucket = new google.storageBucket.StorageBucket(this, 'destinationBucket', {
      lifecycleRule: [{
        action: {
          type: 'delete',
        },
        condition: {
          age: 1,
        },
      }],
      location: region,
      name: `destination-bucket-${project}`,      
    });

    const asset = new TerraformAsset(this, 'asset', {
      path: path.resolve('function'),
      type: AssetType.ARCHIVE,
    });

    const assetObject = new google.storageBucketObject.StorageBucketObject(this, 'assetObject', {
      bucket: assetBucket.name,
      name: asset.assetHash,
      source: asset.path,
    });

    new google.cloudfunctions2Function.Cloudfunctions2Function(this, 'thumbnailMaker', {
      buildConfig: {
        runtime: 'go120',
        source: {
          storageSource: {
            bucket: assetBucket.name,
            object: assetObject.name,
          },
        },
      },
      name: 'thumbnail-maker',
      serviceConfig: {
        environmentVariables: {
          'DESTINATION_BUCKET': destinationBucket.name,
        },
        minInstanceCount: 0,
        maxInstanceCount: 1,
        serviceAccountEmail: functionRunner.email,
      },
      eventTrigger: {
        eventFilters: [{
          attribute: 'bucket',
          value: sourceBucket.name,
        }],
        eventType: 'google.cloud.storage.object.v1.finalized',
      },
    });

    new google.cloudbuildTrigger.CloudbuildTrigger(this, 'buildTrigger', {
      filename: 'cloudbuild.yaml',
      github: {
        owner: 'hsmtkk',
        name: repository,
        push: {
          branch: 'main',
        },
      },
    });

  }
}

const app = new App();
new MyStack(app, "ubiquitous-spoon");
app.synth();
