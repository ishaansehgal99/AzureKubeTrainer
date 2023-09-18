Metaflow Evaluation and Setup Experience

Upon evaluating Metaflow, I found its setup process to be extensive. The tutorial I followed can be found here: Azure Kubernetes Deployment with Metaflow (https://outerbounds.com/engineering/deployment/azure-k8s/deployment/).

Setup Steps and Challenges:

1. Terraform Installation: My first step was to install Terraform. Using it, I provisioned several Azure resources including an Azure Storage Account, a PostgreSQL database, and set up a virtual network.

2. Azure Resource Verification: Post deployment, I ensured that all resources were correctly provisioned and debugged any issues encountered.

3. Service Principal Authentication: The next step was authenticating with Azure using a service principal role.

4. Metaflow CLI Installation Issues: I ran into significant difficulties when attempting to install the Metaflow CLI on my Windows machine with WSL. The installation demanded multiple additional packages as outlined in the official Metaflow documentation. It should be noted that these instructions are primarily tested on Ubuntu 18.04.

A primary challenge was the installation of R. The following error stumped my progress:

```
The following packages have unmet dependencies:
 r-cran-cluster : Depends: r-api-3.5
 r-cran-nlme : Depends: r-api-3.5
 r-cran-rpart : Depends: r-api-3.5
E: Unable to correct problems, you have held broken packages.
```

Despite my efforts in addressing the missing dependencies and configurations, I was unable to overcome this error. Consequently, the R installation failure on WSL prevented me from downloading the Metaflow CLI.

5. Demo Script: In anticipation of a successful CLI setup, I prepared a demo script for testing.

Final Thoughts:

While Metaflow presents undeniable strengths—convenience, simplicity, and power—it also comes with notable drawbacks. The framework can be resource-intensive and bulky. Its design mandates users to adapt their code to fit Metaflow's step architecture and also necessitates hosting a database to support its backend. Such complexities can be daunting and might not appeal to everyone, particularly when seeking flexibility in workflow tools.
