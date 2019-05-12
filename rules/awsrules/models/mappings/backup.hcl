import = "aws-sdk-go/models/apis/backup/2018-11-15/api-2.json"

mapping "aws_backup_selection" {
  name = BackupSelectionName
}

mapping "aws_backup_vault" {
  name = BackupVaultName
}

test "aws_backup_selection" "name" {
  ok = "tf_example_backup_selection"
  ng = "tf_example_backup_selection_tf_example_backup_selection"
}

test "aws_backup_vault" "name" {
  ok = "example_backup_vault"
  ng = "example_backup_vault_example_backup_vault_example_backup_vault"
}
