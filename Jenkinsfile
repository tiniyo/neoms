def CONTAINER_NAME="neoms"
def CONTAINER_TAG="latest"
node {


    stage('Initialize')
  {
      def dockerHome = tool 'mydocker'
      env.PATH = "${dockerHome}/bin:${env.PATH}"
  }
    stage('Checkout') 
    {
        checkout scm
    }


 stage('Build Image'){
        imageBuild(CONTAINER_NAME, CONTAINER_TAG)
    }

    stage('Push to Docker Registry'){
        withCredentials([usernamePassword(credentialsId: 'docker_registry', usernameVariable: 'USERNAME', passwordVariable: 'PASSWORD')]) {
            pushToImage(CONTAINER_NAME, CONTAINER_TAG, USERNAME, PASSWORD)
        }
    }
}


def imageBuild(containerName, tag){
    sh "docker build -t $containerName:$tag  -t $containerName --pull --no-cache ."
    echo "Image build complete"
}

def pushToImage(containerName, tag, dockerUser, dockerPassword){
    sh "docker login -u $dockerUser -p $dockerPassword https://registry.tiniyo.com"
    sh "docker tag $containerName:$tag registry.tiniyo.com/$containerName:$tag"
    sh "docker push registry.tiniyo.com/$containerName:$tag"
    echo "Image push complete"
}
