module.exports = function(grunt) {
  grunt.registerMultiTask("generate-suite", "", function() {
    var _ = require("underscore");
    var done = this.async();

    this.files.forEach(function(file) {
      var output = file.src.map(function(fname) {
        return _.template("require('../<%= fname %>');", { fname: fname });
      }).join("\n");

      grunt.file.write(file.dest, output);
    });

    done();
  });
};
