package main

import (
	"encoding/base64"
	"os"
	"path/filepath"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	ds "github.com/deltastreaminc/pulumi-deltastream/sdk/go/pulumi-deltastream"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		apiKey := os.Getenv("DELTASTREAM_API_KEY")
		server := os.Getenv("DELTASTREAM_SERVER")
		bootstrap := os.Getenv("KAFKA_WITH_CA_URIS")
		saslUser := os.Getenv("KAFKA_WITH_CA_SCRAM_USER")
		saslPass := os.Getenv("KAFKA_WITH_CA_SCRAM_PASS")
		caCertB64 := os.Getenv("KAFKA_WITH_CA_CERT")
		if apiKey == "" || server == "" || bootstrap == "" || saslUser == "" || saslPass == "" || caCertB64 == "" {
			return nil // test harness will skip validation if outputs absent
		}

		certBytes, err := base64.StdEncoding.DecodeString(caCertB64)
		if err != nil {
			return err
		}
		// Write the identical cert bytes to a different fixed path than step3. Since
		// tlsCaCertFile is compared by path, this deterministically forces
		// storeKafkaUpdate to detect a change and re-exercise the CA-cert update path.
		certPath := filepath.Join(os.TempDir(), "ds-kafka-with-ca-v2.pem")
		if err := os.WriteFile(certPath, certBytes, 0o600); err != nil {
			return err
		}

		prov, err := ds.NewProvider(ctx, "deltastream", &ds.ProviderArgs{Server: pulumi.String(server), ApiKey: pulumi.StringPtr(apiKey)})
		if err != nil {
			return err
		}
		store, err := ds.NewStore(ctx, "kafka-store-ca", &ds.StoreArgs{
			Name: pulumi.String("pulumi_kafka_store_ca"),
			Kafka: (ds.KafkaInputsArgs{
				Uris:             pulumi.String(bootstrap),
				SaslHashFunction: pulumi.String("SHA512"),
				SaslUsername:     pulumi.StringPtr(saslUser),
				SaslPassword:     pulumi.StringPtr(saslPass),
				TlsCaCertFile:    pulumi.StringPtr(certPath),
			}).ToKafkaInputsPtrOutput(),
		}, pulumi.Provider(prov))
		if err != nil {
			return err
		}
		ctx.Export("store_auth_mode", pulumi.String("SHA512_CA"))
		ctx.Export("store_state", store.State)
		return nil
	})
}
