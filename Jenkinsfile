pipeline {
  agent any
  options {
    disableConcurrentBuilds()
    buildDiscarder(logRotator(numToKeepStr: '5'))
    timeout(time: 1, unit: 'HOURS')
  }
  stages {
    stage('Setup') {
      steps {
        slackSend channel: "#builds", message: "Build <${env.BUILD_URL}|#${env.BUILD_NUMBER}> of ${env.JOB_NAME} has been started by ${env.GIT_AUTHOR_NAME}"
        sh 'export'
      }
    }
    stage('Build') {
      steps {
        sh 'docker build --tag registry.dreamcove.com/dreamcove/stockfish-server:latest .'
      }
    }
    stage('Publish') {
      when {
        expression {
          env.BRANCH_NAME == 'master'
        }
      }
      steps {
        withCredentials([
          string(credentialsId: 'DOCKER_PASSWORD', variable: 'DOCKER_PASSWORD'),
          string(credentialsId: 'DOCKER_USERNAME', variable: 'DOCKER_USERNAME'),
        ]) {
          sh 'docker login registry.dreamcove.com -u $DOCKER_USERNAME -p $DOCKER_PASSWORD && docker push registry.dreamcove.com/dreamcove/stockfish-server'
        }
      }
    }
  }
  post {
    success {
      slackSend channel: "#builds", message: "Build <${env.BUILD_URL}|#${env.BUILD_NUMBER}> of ${env.JOB_NAME} completed successfully"
    }
    failure {
      slackSend channel: "#builds", message: "Build <${env.BUILD_URL}|#${env.BUILD_NUMBER}> of ${env.JOB_NAME} failed"
    }
  }
}
