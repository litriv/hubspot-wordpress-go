Blog migration tool for migrating content from HubSpot to Wordpress.

Shell scripts retrieve content, parsers parse content into structs and writers generate PHP code, for uploading by the PHP project.

To run the parsers and writers, with some sanity checking:

go run generator.go

# TODO

Remove some client specific code from the sources 