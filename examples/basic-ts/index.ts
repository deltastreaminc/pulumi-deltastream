import * as pulumi from "@pulumi/pulumi";
import * as deltastream from "@deltastream/pulumi";
import { getSystemErrorMap } from "util";


const provider = new deltastream.Provider("deltastream", {
    apiKey: "api-key",
    organization: "org-id",
    server: "server-uri",
});

const db1 = new deltastream.Database("pulumi2", {
    name: "pulumi-ts2",
}, { 
    provider: provider,
 });
