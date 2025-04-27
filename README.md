# Small Language Model Prompt Engineering

## About

An application designed to leverage the efficiency of small language models by implementing prompt engineering, inferencing techniques, and external tooling to increase their accuracy.

We do this by creating what we call pipelines. Pipelines are a way of orchestrating models in a linear fashion to create better answers using multiple models versus one large model.
This is a test to prove that small models (SLMs) can punch above their size and get the job done. Our project is useful in scenarios where there is little memory to work with.

We currently have four pipelines.
1. Simple
2. Chain of Models
3. Debate
4. Embedding

All pipelines, except Embedding, have access to the tools and functionality.

### Simple Pipeline
This pipeline is meant to be used to when only a single language model is desired. This is also benificial if you have a larger model and want to do things that traditional way.

### Chain of Models
This pipeline is meant to be used when you want to emulate a Chain of Thought process. This first model in the pipeline generates initial thoughts and answers the question.
The final model generates the answer that is returned to the client. The intermediate models will operate on the previous answer, either affirming it or correcting it.

### Debate of Models
This pipeline is meant to be used when you want models to debate on a topic. It has been found that debate helps models to generate better answers.
This is especially true when there is expert level knowledge present in the debate which usually challenging to attain when you can't run LLMs. With this pipeline, the SLMs can each
act as a expert.

## Installation

1. We need to install some dependencies so that we can build and run the project. The first thing we need to install is Docker.
Follow your OS specific instructions to install Docker.

2. This project also uses other projects as submodules. To get these projects run the following command to clone the repo and get those dependencies.
```bash
git clone --recursive https://github.com/StoneG24/slape.git
```

3. Create a folder named *models*. SLaPE will create this folder for you, along with checking if that folder exists on startup.
We also download an embedding model for use in the project. [Casual-Autopsy/snowflake-arctic-embed-l-v2.0-gguf](https://huggingface.co/Casual-Autopsy/snowflake-arctic-embed-l-v2.0-gguf)

### GPU Support

After that, if you want to run the containers with a gpu, you'll need to install the nvidia continer toolkit along with the appropriate drivers if needed.
[NVIDIA Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest)

Refer to the proper documentation for rocm.

## Building
This command downloads everything and builds the app for you. A models folder is still needed in the folder that SLaPE is ran in.
```bash
go install github.com/StoneG24/slape@latest
```
## Socket Interactions

### Linux

To run the app you need to turn on the docker socket. This allows the app to talk to the socket and controll its components.

For non-rootful use cases

```bash
sudo systemctl start docker
```

To close the socket on linux,

```bash
sudo systemctl stop docker
```

### Windows

For windows this process is managed by docker desktop.

### Cleanup
SLaPE cleans up its resources. In the event of a crash things may not clean up properly.
To help with this, some commands are included to cleanup those resources. **NOTE** This assumes you are not running any other container setups with docker.
If you are, then clean up the resources on an individual basis.

This command will tell you how much of your disk is currently being used by docker

```bash
docker system df
```

These commands are good for cleaning up these old docker resources. SLaPE should clean these up but, currently, if errors occur it won't.

```bash
docker container prune
```

```bash
docker image prune
```

```bash
docker builder prune
```

## Configuration
To configure the project we used a simple and unorthodox approach.
Instead of using a yaml file, we used a go file that maintains global constants and variables in the program.
These variables are referenced throughtout the program and make it easy to make changes to prompts and numerical values.

these live in the [defs file](pkg/vars/defs.go).

For prompts we also have go files that store the strings that encapsulate the prompts.

these live in [prompts](pkg/prompt/prompt.go).
**NOTE** we also a have set of security prompts for use in demos in [security promtps](pkg/prompt/secprompts.go).

This was choice was made to keep the logic simple and create a binary that could be bundled and moved to remote servers if needed.
A replacement for this would be embedding a yaml file and reading the values at runtime. 
This was originally to complex for the scope of the project.

## Documentation
Our code uses go doc comments as a way of effectively documenting our code.

This tool is included in our tool list of the project.

To run the doc server locally run this command
```bash
godoc -index -notes="BUG|TODO|NOTE"
```
And travel to this url in your browser http://localhost:6060/pkg/github.com/StoneG24/slape/.

## Security
To run security checks on the repo run these commands,

```bash
go vet ./...
go tool staticcheck ./...
go tool govulncheck
```
All of these binaries should have been installed with golang and this package.

## Features
We currently have several feautures to aid in the improvement of SLMs.
We have optional Internet Search, along with an extra thinking step.

### Thinking
This is currently a prototype of a thought process meant to give the model extra time to consider characteristics and behaviors of a given problem.
To enable this, pass in a "thinking":"1" into your json request to our endpoints.

### Internet Search
This is another optional prototype. It is meant to give a model access to the internet for updated information compared to what it was trained on.
It should be noted that the model itself does not make the request. It merely generates the guery used to search the web. The rest is handled internally.

### Function Calling (WIP)

### Indexing RAG (LightRag/MiniRag) (WIP)
The code is present for guerying the database but it is untested and not integrated into the context.

## Reference

Here are some of the research papers that we used to aid us in development.
