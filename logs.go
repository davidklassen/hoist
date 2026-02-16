package main

import "strconv"

func dockerLogsArgs(container, since string, n int, follow bool) []string {
	args := []string{"logs"}

	if n > 0 {
		args = append(args, "--tail", strconv.Itoa(n))
	}

	if since != "" {
		args = append(args, "--since", since)
	}

	if follow {
		args = append(args, "-f")
	}

	args = append(args, container)
	return args
}

func joinArgs(args []string) string {
	result := ""
	for i, a := range args {
		if i > 0 {
			result += " "
		}
		result += a
	}
	return result
}
