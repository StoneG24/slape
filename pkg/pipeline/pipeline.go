/*
Package pipline is meant to help facilitate the running of several models in sequential order.

For machines with small amounts of resources the pipelines will have to manage models by orchestrating models as need be.
Pipelines also dictate the structure of conversations between the models. See the following docs for more info.
All pipelines are based off of prompting techniques that are used in the realm of Test Time Compute(TTC) optimization.
*/
package pipeline

type Pipeline interface {
	Setup() error
	Generate() (string, error)
	Shutdown() error
}
