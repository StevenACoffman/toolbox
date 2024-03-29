// This is a a javascript (or similar language) file

// sync-start:correct 249234014 gons.go
const someCode = "does a thing";
console.log(someCode);
// sync-end:correct

const outputDir = path.join(rootDir, "genwebpack/khan-apollo");

// Handle auto-building the GraphQL Schemas + Fragment Types
module.exports = new FileWatcherPlugin({
    name: "GraphQL Schema Files",

    // The location of the Python files that are used to generate the
    // schema
    filesToWatch: [
        path.join(rootDir, "*/graphql/**/*.py"),
        path.join(rootDir, "*/*/graphql/**/*.py"),
    ],
