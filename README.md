
# Issues (Eric Schmidt/Crossbeam fork)

## open

1. Need a wait after api key creation so that we don't try to use an api key before its ready.

2. Deletes not implemented for APIKey, or Schema Registry.  Not critical for our use case, these deletes get taken care of on sa and env delete.

3. Need to DRY out all the new resources, they can share a common HTTP client and much of their config, ex: Auth and Content type headers. 

4. Error handling: permissions on topic create

## resolved

~~Need to allow credentialed operations to use ambient credentials: create topic, create acls, etc.~~ Least impact is to use new api-key resource to generate an api-key and use that in the credentials blocks for the required resources.

~~KSQLDB not yet implemented~~

~~API Keys are not properly associated with provided service account, they are associating with the SA that runs terraform instead. It appears that the call needs the numeric user id for the desired SA to be correctly populated, but the SA doesn't expose this id... so we need to hack that in, likely with a custom data source for SA... this appears to be deprecated functionality that is still required for this particular operation and may be part of the delay in improvements/releases to this provider.~~ conflunet did provide a nice utility function for converting temporarily.

# Terraform Provider for Confluent Cloud

The Terraform Confluent Cloud provider is a plugin for Terraform that allows for the lifecycle management of Confluent Cloud resources.
This provider is maintained by Confluent.

## Quick Starts

- [Running an example](docs/guides/sample-project.md)
- [Developing the provider](docs/DEVELOPING.md)

## Documentation

Full documentation is available on the [Terraform website](https://registry.terraform.io/providers/confluentinc/confluentcloud/latest/docs).

## License

Copyright 2021 Confluent Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
