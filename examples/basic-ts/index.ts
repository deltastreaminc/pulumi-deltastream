import * as pulumi from "@pulumi/pulumi";
import * as deltastream from "@deltastream/pulumi";
import { randomUUID } from "crypto";

const config = new pulumi.Config();
const apiKey = config.require("DS_API_KEY");
const orgID = config.require("DS_ORGANIZATION_ID");
const serverUri = config.require("DS_SERVER_URI");

const provider = new deltastream.Provider("deltastream", {
    apiKey: apiKey,
    organization: orgID,
    server: serverUri,
});

const dbName = pulumi.interpolate`pulumi-ts-db-${randomUUID()}`;

(async() => {
    const db = new deltastream.Database("db", {
        name: dbName,
    }, { 
        provider: provider,
    });
    const dbOutputName = (await db.name).get()

    const dbList = await deltastream.getDatabases({
        provider: provider,
    });

    let dbExists = false;
    for (const db of dbList.items) {
        if (dbOutputName === db.name) {
            dbExists = true;
        }
    }
    if (!dbExists) {
        throw new Error(`Database ${dbName} does not exist.`);
    }
})();
