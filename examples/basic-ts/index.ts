import * as pulumi from "@pulumi/pulumi";
import * as deltastream from "@deltastream/pulumi-deltastream";
import { randomUUID } from "crypto";
import { awaitStackRegistrations } from "@pulumi/pulumi/runtime";

const config = new pulumi.Config();
const apiKey = config.require("DS_API_KEY");
const orgID = config.require("DS_ORGANIZATION_ID");
const serverUri = config.require("DS_SERVER_URI");

export = async () => {
    const provider = new deltastream.Provider("deltastream", {
        apiKey: apiKey,
        organization: orgID,
        server: serverUri,
    });

    const db = new deltastream.Database("db", {
        name: `pulumi-ts-database`,
    }, { 
        provider: provider,
    })
}
