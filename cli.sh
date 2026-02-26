#!/bin/bash

function use1() {
    cleanText="$1 $2$3$4$5$6$7$8$9${10}${11}${12}${13}${14}${15}${16}${17}${18}${19}${20}"
    echo -e "\033[1;31m            Use:\033[0m $cleanText"
}

function expected1() {
  cleanText="$1 $2$3$4$5$6$7$8$9${10}${11}${12}${13}${14}${15}${16}${17}${18}${19}${20}"
  echo -e "\033[1;31m       Expected:\033[0m $cleanText"
}

function execute(){
  ONE_COMMAND="$1"
  use1 "$ONE_COMMAND"
  $ONE_COMMAND
}

if [ $1 == "health" ]; then
  execute "curl --location http://localhost:8080/health"
elif [ $1 == "select" ]; then
  echo ""
  if [ -z "$2" ]; then
    execute "curl --location http://localhost:8080/select/users"
  else
    execute "curl --location http://localhost:8080/select/users?where=name%3D%27$2%27"
  fi
  echo ""
  expected1 "cli select [<name>]"
elif [ $1 == "insert" ]; then
  if [ -z "$2" ]; then
    expected1 "cli insert <name>"
    exit 0
  fi
  echo ""
  curl --location 'http://localhost:8080/insert/users' \
           --header 'Content-Type: application/json' \
           --data-raw '{"name": "'$2'", "email": "'$2'@example.com"}'
  echo ""
  echo ""
fi
expected1 "cli [health | select | insert]"
exit 0