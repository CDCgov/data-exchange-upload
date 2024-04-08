package hooks

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/cdcgov/data-exchange-upload/upload-server/experimental/metadatav1"
)

func hydrateManifestV1(manifest *map[string]*string, hydrateV1Config metadatav1.HydrateV1Config) {

	// Add use-case specific fields and their values.
	for _, field := range hydrateV1Config.MetadataConfig.Fields {

		if field.FieldName == "" {
			continue
		} // .if

		// continue if field already provided.
		if _, exists := (*manifest)[field.FieldName]; exists {
			continue
		} // .if

		if field.DefaultValue != "" {
			(*manifest)[field.FieldName] = to.Ptr(field.DefaultValue)
			continue
		} // .if

		if field.CompatFieldName != "" {
			(*manifest)[field.FieldName] = (*manifest)[field.CompatFieldName]
			continue
		} // .if

		(*manifest)[field.FieldName] = to.Ptr("")
	} // .for

} // .hydrateMetadataV1
