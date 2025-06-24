pipeline {
    agent any

    environment {
        DOCKER_IMAGE_NAME = "app"
        DOCKER_IMAGE_TAG = "${env.BRANCH_NAME}-${env.BUILD_NUMBER}"
        DOCKER_TAR_FILE = "${DOCKER_IMAGE_NAME}-${DOCKER_IMAGE_TAG}.tar"

        BRANCH_DIR = "/home/ubuntu/sudarshan/ride-sharing-notification/${env.BRANCH_NAME}/"
        DEPLOYMENT_TIMEOUT = "10m"
        HEALTH_CHECK_TIMEOUT = "2m"
    }

    stages {
        stage('Verify Environment') {
            steps {
                script {
                    def envName = (env.BRANCH_NAME ?: "dev").toUpperCase()
                    env.envFileCredentialId = "RIDE_NOTIFICATION_${envName}_ENV"

                    echo """
                    ===== Deployment Configuration =====
                    Environment: ${envName}
                    Branch: ${env.BRANCH_NAME ?: 'dev'}
                    Build Number: ${env.BUILD_NUMBER}
                    Image Tag: ${env.DOCKER_IMAGE_TAG}
                    Target Directory: ${env.BRANCH_DIR}
                    """
                }
            }
        }

        stage('Build Docker Image') {
            steps {
                script {
                    withCredentials([file(credentialsId: env.envFileCredentialId, variable: 'ENV_FILE')]) {
                        sh '''
                            #!/bin/bash
                            set -e

                            echo "===== Preparing environment ====="
                            install -m 600 "$ENV_FILE" .env

                            echo "===== Building Docker image ====="
                            docker build --no-cache -t ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} .

                            echo "===== Saving Docker image ====="
                            docker save -o ${DOCKER_TAR_FILE} ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}

                            echo "===== Verification ====="
                            ls -la ${DOCKER_TAR_FILE}
                            ls -la .env
                        '''
                    }
                }
            }
        }

        stage('Deploy to Server') {
            steps {
                script {
                    withCredentials([
                        string(credentialsId: 'HOST_IP', variable: 'SERVER_HOST'),
                        string(credentialsId: 'SERVER_USER', variable: 'SERVER_USER'),
                        file(credentialsId: 'SERVER_KEY', variable: 'SSH_KEY_PATH'),
                        string(credentialsId: 'SERVER_PORT', variable: 'SSH_PORT')
                    ]) {
                        sh """
                            # Transfer files
                            scp -i "$SSH_KEY_PATH" -P "$SSH_PORT" \\
                                ${DOCKER_TAR_FILE} docker-compose.yml .env \\
                                ${SERVER_USER}@${SERVER_HOST}:${BRANCH_DIR}

                            # Execute deployment
                            ssh -i "$SSH_KEY_PATH" -p "$SSH_PORT" -tt ${SERVER_USER}@${SERVER_HOST} "
                                set -e
                                cd ${BRANCH_DIR}

                                # Set permissions
                                chmod 600 .env
                                chmod 644 docker-compose.yml

                                # Load image
                                echo '===== Loading Docker image ====='
                                docker load -i ${DOCKER_TAR_FILE}

                                # Start services
                                echo '===== Starting services ====='
                                docker-compose down || true
                                DOCKER_IMAGE_TAG=${DOCKER_IMAGE_TAG} docker-compose up -d --remove-orphans

                                # Cleanup
                                rm -f ${DOCKER_TAR_FILE}
                                echo '===== Deployment complete ====='
                            "
                        """
                    }
                }
            }
        }
    }
}