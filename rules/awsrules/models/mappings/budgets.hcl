import = "aws-sdk-go/models/apis/budgets/2016-10-20/api-2.json"

mapping "aws_budgets_budget" {
  account_id  = AccountId
  name        = BudgetName
  budget_type = BudgetType
  time_unit   = TimeUnit
}

test "aws_budgets_budget" "account_id" {
  ok = "123456789012"
  ng = "abcdefghijkl"
}

test "aws_budgets_budget" "name" {
  ok = "budget-ec2-monthly"
  ng = "budget:ec2:monthly"
}

test "aws_budgets_budget" "budget_type" {
  ok = "USAGE"
  ng = "MONEY"
}

test "aws_budgets_budget" "time_unit" {
  ok = "MONTHLY"
  ng = "HOURLY"
}
