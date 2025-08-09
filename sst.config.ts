import { SSTConfig } from "sst";
import { Api } from "sst/constructs";
import { Function } from "sst/constructs";

export default {
  config(_input) {
    return {
      name: "lux-api",
      region: "us-east-1",
    };
  },
  stacks(app) {
    app.stack(function Site({ stack }) {
      // Create API Gateway
      const api = new Api(stack, "LuxAPI", {
        cors: {
          allowCredentials: true,
          allowHeaders: ["*"],
          allowMethods: ["GET", "POST", "OPTIONS"],
          allowOrigins: ["*"],
        },
      });

      // Create Lambda function for video extraction
      const lambdaFunction = new Function(stack, "LuxExtractor", {
        url: true,
        runtime: "provided.al2023",
        handler: "bootstrap",
        architecture: "arm64",
        timeout: "30 seconds",
        memory: "1024 MB",
        environment: {
          GOOS: "linux",
          GOARCH: "arm64",
        },
      });

      // Add routes to API Gateway
      api.route("GET /extract", {
        function: {
          handler: lambdaFunction,
        },
      });

      api.route("POST /extract", {
        function: {
          handler: lambdaFunction,
        },
      });

      api.route("OPTIONS /extract", {
        function: {
          handler: lambdaFunction,
        },
      });

      // Output the API URL
      stack.addOutputs({
        apiUrl: api.url,
      });
    });
  },
} satisfies SSTConfig;
