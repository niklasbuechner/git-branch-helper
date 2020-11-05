package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type Branch struct {
	Name string
	RemoteBranch string
	RemoteStatus string
}

func main() {
    currentBranch := strings.Trim(execCommand("git", "branch", "--show-current"), "\n \t\r")

    fmt.Println("Fetching all references")
    fmt.Println("")
	execCommand("git", "fetch")

	localBranches := execCommand("git", "branch", "-vv")
	branchesWithGithubDifsRegex, _ := regexp.Compile("^[ *]{2}([\\S]+)[\\s]+[\\S]+[\\s]+\\[([\\S]+): (.+)]")
    branchesWithoutRemoteBranchesRegex, _ := regexp.Compile("^[ *]{2}([\\S]+)[\\s]+[\\S]+[\\s]+[^[]")

	var branches []Branch
	for _, localBranch := range strings.Split(string(localBranches), "\n") {
		branch := branchesWithGithubDifsRegex.FindStringSubmatch(localBranch)
		if branch != nil {
			branches = append(branches, Branch{
				Name: string(branch[1]),
				RemoteBranch: string(branch[2]),
				RemoteStatus: string(branch[3]),
			})

			continue
		} else {
            branch := branchesWithoutRemoteBranchesRegex.FindStringSubmatch(localBranch)
            if branch != nil {
                branches = append(branches, Branch{
                    Name: string(branch[1]),
                    RemoteBranch: "",
                    RemoteStatus: "",
                })
            }
        }
	}

    var branchesToDelete []Branch
    var branchesWithoutRemoteBranch []Branch
	for _, branch := range branches {
		if branch.RemoteStatus == "gone" {
		    branchesToDelete = append(branchesToDelete, branch)
		} else if branch.RemoteBranch == "" {
            branchesWithoutRemoteBranch = append(branchesWithoutRemoteBranch, branch)
        }
	}

	if len(branchesToDelete) == 0 {
		fmt.Println("No branches are stale")
	} else {
        listBranches("Branches to delete because their remote counterparts are gone:", branchesToDelete)

        input := getInput("Would you like to delete these branches? [y/n]")
        if input == "y\n" {
            for _, branchToDelete := range branchesToDelete {
                if branchToDelete.Name == currentBranch {
                    currentBranch = getMainBranch(branches)
                    execCommand("git", "checkout", currentBranch)
                }

                fmt.Printf("Deleteing branch %s...\n", branchToDelete.Name)
                execCommand("git", "branch", "-D", branchToDelete.Name)
            }
        } else {
            fmt.Println("Branches where NOT deleted.")
        }
    }

    if len(branchesWithoutRemoteBranch) == 0 {
        fmt.Println("All branches have corresponding remote branches");
    } else {
        listBranches("Branches without remote counterparts:", branchesWithoutRemoteBranch)

        input := getInput("Would you like create these branches in your origin repository? [y/n]")
        if input == "y\n" {
            for _, branchToPush := range branchesWithoutRemoteBranch {
                fmt.Printf("Creating branch %s...\n", branchToPush.Name)
                execCommand("git", "checkout", branchToPush.Name)
                execCommand("git", "push", "-u", "origin", branchToPush.Name)
            }

            execCommand("git", "checkout", currentBranch)
        } else {
            fmt.Println("Branches where NOT pushed.")
        }
    }

	fmt.Println("Done")
}

func execCommand(cmd string , arguments ...string) string {
    command := exec.Command(cmd, arguments...)
	output, err := command.CombinedOutput()
	if err != nil {
        fmt.Printf("Cmd error %s", string(output))
        panic("")
    }

    return string(output)
}

func listBranches(heading string, branches []Branch) {
    fmt.Println(heading)
    for _, branch := range branches {
        fmt.Printf("- %s\n", branch.Name)
    }
}

func getInput(question string) string {
    fmt.Printf("\n%s ", question)
	stdinReader := bufio.NewReader(os.Stdin)
    input, _ := stdinReader.ReadString('\n')
    fmt.Println("")

    return input
}

func getMainBranch(branches []Branch) string {
    mainBranch := ""
    for _, branch := range branches {
        if branch.Name == "main" {
            return branch.Name
        } else if branch.Name == "master" {
            mainBranch = branch.Name
        }
    }

    return mainBranch
}
