package jira

func NewSTIPClient() (*JiraClient, error) {
	jiraClient := JiraClient{}
	err := jiraClient.Authenticate("luis.santos@lab900.com", "ATATT3xFfGF0CC7CnjGgQ1jlJs0u7efZJ4Bb8wFCaPAPUsIuUSW-AYMIlkjcSSGiJqrVlNkYib_OWJI34WY0hR63nCG2tUZhmEWgg-6BX2VhprIvBEcHh5lWsbtXo8huqqlrveIbazAylU2vxvDa7MA2VbJ-K41-3PIg9yCqKEe7C7-17Ac3tD8=4D9ECB3A", "https://sea-tank.atlassian.net/")
	return &jiraClient, err
}
