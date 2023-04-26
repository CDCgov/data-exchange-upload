az storage blob upload --account-name $1 --container-name tusd-file-hooks --account-key $2 --name tus/file-hooks/metadata-verify/allowed_destination_and_events.json --file allowed_destination_and_events.json --overwrite
az storage blob upload --account-name $1 --container-name tusd-file-hooks --account-key $2 --name tus/file-hooks/metadata-verify/ndlp-ri-meta-definition.json --file ndlp-ri-meta-definition.json --overwrite
az storage blob upload --account-name $1 --container-name tusd-file-hooks --account-key $2 --name tus/file-hooks/metadata-verify/pre-create --file pre-create --overwrite
az storage blob upload --account-name $1 --container-name tusd-file-hooks --account-key $2 --name pre-create-bin --file pre-create-bin --overwrite
