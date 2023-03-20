# Cheesecake

## Getting Started

To get started with the project, follow the instructions below to set up both the model and the frontend.

### Prerequisites

Ensure that you have the following software installed on your system:

1. [Julia](https://julialang.org/downloads/) (for the model)
2. [Node.js](https://nodejs.org/en/download/) (for the frontend)
3. [Yarn](https://yarnpkg.com/getting-started/install) (for the frontend)

### Model Setup

The `model` directory contains a Julia project for generating predictions using the Bayesian model. To set up and run the model, follow these steps:

1. Navigate to the `model` directory:

   ```bash
   cd model
   ```

2. Add your key from [The Blue Alliance](https://www.thebluealliance.com/apidocs).

   ```bash
   echo 'KEY="<key>"' >> python/.env
   ```

3. Run the data fetching script.

   ```bash
   cd python && python 2023.py && cd ..
   ```

4. Start the Julia REPL:

   ```bash
   julia
   ```

5. Activate the DrWatson project environment and install the required dependencies:

   ```julia
   using DrWatson
   @quickactivate
   import Pkg; Pkg.instantiate()
   ```

6. Run the model:

   ```julia
   include("scripts/run_frc2023.jl")
   ```

### Frontend Setup

The `frontend` directory contains a React app for visualizing the predictions generated by the model. To set up and run the frontend, follow these steps:

1. Navigate to the `frontend` directory:

   ```bash
   cd frontend
   ```

2. Install the required dependencies using Yarn:

   ```bash
   yarn install
   ```

3. Start the development server:

   ```bash
   yarn run dev
   ```

The React app should now be running at \`http://localhost:3000\` in your web browser.
