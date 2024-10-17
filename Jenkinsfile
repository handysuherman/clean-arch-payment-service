@Library(value='ci-shared-library@main', changelog=false) _

import slackUtils.slackMessage.ProductionChannel
import slackUtils.slackMessage.DevelopmentChannel
import deploymentUtils.Production

pipeline {
    agent any

    environment {
        APP_ID = '81d6fa93d9fe88c6148c-payment'
        APP_NAME = "wisata-desa-payment-backend"
        GIT_RELEASES = "https://github.com/wikanproductions/${env.APP_NAME}/releases"
        GIT_VERSION = sh(returnStdout: true, script: 'git fetch --tags && git tag --points-at HEAD | awk NF').trim()
        GIT_COMMIT_MSG = sh(script: 'git log -1 --pretty=%B ${GIT_COMMIT} | head -n1', returnStdout: true).stripIndent().trim()
        GIT_AUTHOR = sh(script: 'git log -1 --pretty=%ae ${GIT_COMMIT} | awk -F "@" \'{print $1}\' | grep -Po "[a-z]{1,}" | head -n1', returnStdout: true).trim()
        
        DOCKER_REGISTRY_ENDPOINT = "docker-registry.local:24115"
        DOCKER_IMAGE_NAME = "${env.APP_NAME}"
        DOCKERFILE_PATH = "build/docker/Dockerfile"
    
        // CERTS
        TLS_ROOT_PATH = '~/certs/production/wikanproductions'

        TLS_APP_LOCAL_PATH = './tls/app'
        TLS_APP_PRODUCTION_PATH = "${env.TLS_ROOT_PATH}/app"

        TLS_ETCD_LOCAL_PATH = './tls/etcd'
        TLS_ETCD_PROD_PATH = "${env.TLS_ROOT_PATH}/etcd/wikan"
        
        TLS_KAFKA_LOCAL_PATH = './tls/kafka'
        TLS_KAFKA_PROD_PATH = "~/certs/production/wikan-general/kafka"

        TLS_REDIS_LOCAL_PATH = './tls/redis'
        TLS_REDIS_PROD_PATH = "${env.TLS_ROOT_PATH}/redis"

        TLS_PQSQL_LOCAL_PATH = './tls/pqsql'
        TLS_PQSQL_PROD_PATH = "${env.TLS_ROOT_PATH}/pqsql"

        TLS_CONSUL_LOCAL_PATH = './tls/consul'
        TLS_CONSUL_PROD_PATH = "${env.TLS_ROOT_PATH}/consul"

        // SUBSCRIPTIONS
        SUBSCRIPTION_ROOT_PATH = '~/etcd-config/wikanproductions'

        SUBSCRIPTION_ETCD_CONFIG_PATH = "${env.SUBSCRIPTION_ROOT_PATH}/wikan"
        SUBSCRIPTION_KAFKA_CONFIG_PATH = '~/etcd-config/general/kafka'
       
        // SLACK
        SLACK_PRODUCTION_CHANNEL = "production-deployment-notifications"
        SLACK_DEVELOPMENT_CHANNEL = "develop-deployment-notifications"
        SLACK_TEAM_DOMAIN = "wikanproduction"
        SLACK_PRODUCTION_CREDENTIAL_ID = "production-slack-token"
        SLACK_DEVELOPMENT_CREDENTIAL_ID = "development-slack-token"
    }

    options {
        disableConcurrentBuilds()
        buildDiscarder(logRotator(numToKeepStr: '3', daysToKeepStr: '3'))
        timestamps()
    }

    triggers {
        pollSCM("* * * * *")
    }

    stages {
        stage('MAIN BRANCH') {
            when {
                branch 'main'
            }

            stages {
                stage("statuses") {
                    steps {
                        script {
                            ProductionChannel.sendStatusMessage(this)
                        }
                    }
                }
                stage("tests: tests app") {
                    steps {
                       script {
                           Production.fetchCerts(this, "${env.TLS_ETCD_PROD_PATH}/service", "${env.TLS_ETCD_LOCAL_PATH}")

                           Production.runBackendUnitTests(this)
                        }
                    }
                }
                stage("build: docker images") {
                    steps {
                        script {
                            Production.buildDockerImageWithoutCheck(this)
                        }
                    }
                }
                stage("deploy: deploying app to server") {
                    steps {
                        script {
                            prepareCerts()
                            sh 'cp ./etcd-config.yaml ./deployments/ansible/roles/constantinopel/files/'
                            sh "cp ${env.TLS_PQSQL_LOCAL_PATH}/ca-cert.pem ./deployments/ansible/roles/constantinopel/files/pq-ca-cert.pem"
                            sh "cp ${env.TLS_PQSQL_LOCAL_PATH}/client-cert.pem ./deployments/ansible/roles/constantinopel/files/pq-client-cert.pem"
                            sh "cp ${env.TLS_PQSQL_LOCAL_PATH}/client-key.pem ./deployments/ansible/roles/constantinopel/files/pq-client-key.pem"
                            sh 'cp ./configs/monitoring/prometheus.yml ./deployments/ansible/roles/constantinopel/files/'
                            sh 'ansible-playbook ./deployments/ansible/playbook.yml -i ./deployments/ansible/inventory --tags "app"'
                        }
                    }
                }
                stage("subscribe-configurations") {
                    steps {
                        script {
                            sh "mv ./configs/etcd/e-prod-config-cli.yaml ./configs/etcd/config-cli.yaml"
                            
                            Production.subscribeConfig(this, "${env.SUBSCRIPTION_ETCD_CONFIG_PATH}")
                            
                            Production.subscribeConfig(this, "${env.SUBSCRIPTION_KAFKA_CONFIG_PATH}")
                        }
                    }
                }
            }
        }
    }

    post {
        cleanup {
            cleanWorkspaceDirs()
        }
    }
}


def prepareCerts() {
    Production.generateAppCerts(this)
    
    Production.fetchServerCerts(this, "${env.TLS_APP_PRODUCTION_PATH}/${env.APP_ID}/service", "${env.TLS_APP_LOCAL_PATH}")

    Production.fetchCerts(this, "${env.TLS_KAFKA_PROD_PATH}/service", "${env.TLS_KAFKA_LOCAL_PATH}")
    
    Production.fetchCerts(this, "${env.TLS_ETCD_PROD_PATH}/service", "${env.TLS_ETCD_LOCAL_PATH}")
    
    Production.fetchCerts(this, "${env.TLS_PQSQL_PROD_PATH}/${env.APP_ID}/service", "${env.TLS_PQSQL_LOCAL_PATH}")
    
    Production.fetchCerts(this, "${env.TLS_REDIS_PROD_PATH}/${env.APP_ID}/service", "${env.TLS_REDIS_LOCAL_PATH}")
    
    Production.fetchCerts(this, "${env.TLS_CONSUL_PROD_PATH}/${env.APP_ID}/service", "${env.TLS_CONSUL_LOCAL_PATH}")
    sh "~/go/bin/etconf --config-file=configs/etcd/e-prod-config-cli.yaml"
}

def handleBuildError(err) {
    if ('SYSTEM' == err.getCauses()[0].getUser().toString()) {
        echo 'timeouts'
    } else {
        echo 'its aborted'
    }
}

def cleanWorkspaceDirs() {
    cleanWs deleteDirs: true, cleanWhenAborted: true, cleanWhenFailure: true, cleanWhenSuccess: true
    dir("${workspace}@tmp") {
        deleteDir()
    }
    dir("${workspace}@libs") {
        deleteDir()
    }
}
