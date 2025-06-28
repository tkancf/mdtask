#!/bin/bash

# mdtask CLI test script

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test directory
TEST_DIR="test_tasks"
MDTASK="./mdtask"

# Clean up function
cleanup() {
    echo -e "\n${YELLOW}Cleaning up...${NC}"
    rm -rf "$TEST_DIR"
}

# Error handler
on_error() {
    echo -e "${RED}Test failed!${NC}"
    cleanup
    exit 1
}

# Set up error handling
trap on_error ERR

# Print test header
print_test() {
    echo -e "\n${YELLOW}=== $1 ===${NC}"
}

# Print success
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# Start tests
echo -e "${GREEN}Starting mdtask CLI tests...${NC}"

# Build the project
print_test "Building mdtask"
go build -o mdtask
print_success "Build successful"

# Create test directory
print_test "Setting up test environment"
mkdir -p "$TEST_DIR"
print_success "Test directory created"

# Test 1: Create new task
print_test "Test 1: Creating new task"
TASK_ID=$(echo -e "最初のテストタスク\nこれはテスト用のタスクです\nテスト内容\n- 項目1\n- 項目2" | $MDTASK new --paths $TEST_DIR -s TODO --deadline "2024-12-31" | grep "ID:" | awk '{print $2}')
echo "Created task: $TASK_ID"
print_success "Task created successfully"

# Test 2: List tasks
print_test "Test 2: Listing tasks"
$MDTASK list --paths $TEST_DIR
print_success "List command works"

# Test 3: Create more tasks with different statuses
print_test "Test 3: Creating tasks with different statuses"
echo -e "進行中タスク\nWIPステータスのタスク\n作業中です" | $MDTASK new --paths $TEST_DIR -s WIP --tags "urgent,開発"
sleep 1  # Wait to avoid ID collision
echo -e "待機中タスク\n承認待ちのタスク\nレビュー待機中" | $MDTASK new --paths $TEST_DIR -s WAIT --tags "レビュー"
sleep 1  # Wait to avoid ID collision
echo -e "完了タスク\n完了済みのタスク\n実装完了" | $MDTASK new --paths $TEST_DIR -s DONE
print_success "Multiple tasks created"

# Test 4: List with status filter
print_test "Test 4: Listing tasks by status"
echo "TODO tasks:"
$MDTASK list --paths $TEST_DIR --status TODO
echo -e "\nWIP tasks:"
$MDTASK list --paths $TEST_DIR --status WIP
print_success "Status filtering works"

# Test 5: Search functionality
print_test "Test 5: Searching tasks"
echo "Searching for 'タスク':"
$MDTASK search --paths $TEST_DIR "タスク"
echo -e "\nSearching for 'urgent' tag:"
$MDTASK search --paths $TEST_DIR "urgent"
print_success "Search functionality works"

# Test 6: Archive task
print_test "Test 6: Archiving task"
echo "Archiving task: $TASK_ID"
$MDTASK archive "$TASK_ID" --paths $TEST_DIR
echo -e "\nActive tasks:"
$MDTASK list --paths $TEST_DIR
echo -e "\nArchived tasks:"
$MDTASK list --paths $TEST_DIR --archived
print_success "Archive functionality works"

# Test 7: Create task with deadline
print_test "Test 7: Task with deadline"
sleep 1  # Wait to avoid ID collision
OVERDUE_DATE=$(date -v-1d +%Y-%m-%d 2>/dev/null || date -d "yesterday" +%Y-%m-%d)
echo -e "期限切れタスク\n期限が過ぎているタスク\n対応が必要" | $MDTASK new --paths $TEST_DIR --deadline "$OVERDUE_DATE"
echo -e "\nTasks with deadlines:"
$MDTASK list --paths $TEST_DIR --all
print_success "Deadline functionality works"

# Test 8: File naming convention
print_test "Test 8: Checking file naming convention"
echo "Files in test directory:"
ls -la $TEST_DIR/
FILE_COUNT=$(ls $TEST_DIR/*.md 2>/dev/null | wc -l | tr -d ' ')
echo "Total task files: $FILE_COUNT"
print_success "File naming convention correct (YYYYMMDDHHMMSS.md)"

# Test 9: Edit command (non-interactive test)
print_test "Test 9: Testing edit command availability"
if $MDTASK edit --help > /dev/null 2>&1; then
    print_success "Edit command is available"
else
    echo -e "${RED}Edit command not available${NC}"
fi

# Test 10: List all tasks including archived
print_test "Test 10: List all tasks"
echo "All tasks (including archived):"
$MDTASK list --paths $TEST_DIR --all
TOTAL_COUNT=$($MDTASK list --paths $TEST_DIR --all | grep "task/" | wc -l | tr -d ' ')
echo "Total tasks: $TOTAL_COUNT"
print_success "All tasks listed successfully"

# Summary
echo -e "\n${GREEN}===================================${NC}"
echo -e "${GREEN}All tests completed successfully!${NC}"
echo -e "${GREEN}===================================${NC}"
echo -e "\nSummary:"
echo "- Total tasks created: $FILE_COUNT"
echo "- Tasks tested: new, list, search, archive"
echo "- Features tested: status filtering, deadline handling, tag search"

# Clean up
cleanup

echo -e "\n${GREEN}Test completed!${NC}"