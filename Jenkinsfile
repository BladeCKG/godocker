// Jenkinsfile (Declarative Pipeline)
/* Requires the Docker Pipeline plugin */
pipeline {
    agent { docker { image 'golang:1.23.1-alpine3.20' } }
    stages {
        stage('build') {
            steps {
                echo 'build!!!'
                sh 'go version'
            }
        }
    }
}