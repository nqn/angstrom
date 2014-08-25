module.exports = function(grunt) {

  [ "grunt-contrib-watch",
    "grunt-browserify",
    "grunt-contrib-copy",
    "grunt-contrib-connect",
    "grunt-contrib-less",
    "grunt-contrib-jshint",
    "grunt-contrib-uglify",
    "grunt-mocha",
    "grunt-react"
  ].forEach(function(n) {
    grunt.loadNpmTasks(n);
  });

  // Add the generate-suite task to grunt.
  require("./plugins/generateSuite")(grunt);

  grunt.initConfig({
    pkg: grunt.file.readJSON("package.json"),
    conf: {
      app: "<%= conf.src %>/**/*.js*",
      compiled: "<%= conf.pub %>/js/<%= pkg.name %>.js",
      tests: [ "tests/**/*.js" ],
      sys: [ "Gruntfile.js", "package.json" ],
      less: "<%= conf.src %>less/",
      css: "<%= conf.pub %>css/",
      src: "src/",
      js: "<%= conf.src %>js/",
      pub: "public/",
      build: "build/"
    },
    less: {
      dev: {
        options: {
          strictMath: true,
          sourceMap: true,
          outputSourceFiles: true,
          sourceMapURL: "<%= pkg.name %>.css.map",
          sourceMapFilename: "<%= conf.css %><%= pkg.name %>.css.map"
        },
        files: {
          "<%= conf.css %>angstrom.css": "<%= conf.less %>angstrom.less"
        }
      }
    },
    copy: {
      html: {
        files: [
          {
            cwd: "<%= conf.src %>",
            src: "**/*.html",
            dest: "<%= conf.pub %>",    // destination folder
            expand: true,           // required when using cwd
            filter: "isFile"
          }
        ]
      },
      img: {
        files: [
          {
            cwd: "<%= conf.src %>",
            src: ["img/**"],
            dest: "<%= conf.pub %>",    // destination folder
            expand: true           // required when using cwd
          }
        ]
      },
      fonts: {
        files: [
          {
            cwd: "<%= conf.src %>",
            src: ["fonts/**"],
            dest: "<%= conf.pub %>",    // destination folder
            expand: true           // required when using cwd
          }
        ]
      },
      css: {
        files: [
          {
            cwd: "<%= conf.less %>",
            src: "**/*.css",
            dest: "<%= conf.css %>",    // destination folder
            expand: true,           // required when using cwd
            filter: "isFile"
          }
        ]
      }
    },
    watch: {
      app: {
        files: [
          "<%= conf.sys %>",
          "<%= conf.src %>"
        ],
        tasks: ["dev"]
      }
      // test: {
      //   files: [
      //     "<%= conf.sys %>",
      //     "<%= conf.tests %>",
      //     "<%= conf.app %>"
      //   ],
      //   tasks: [ "runtests" ]
      // }
    },
    browserify: {
      app: {
        src: [ "<%= conf.js %>/<%= pkg.main %>" ],
        dest: "<%= conf.compiled %>",
        options: {
          bundleOptions: {
            insertGlobalVars: {
              /* Insert a `process` global that imitates the production Node
               * environment required by React. This prevents full file paths
               * from being output to the compiled JS.
               *
               * TODO(ssorallen): Remove this once pull request #31[1] is merged
               * in the `insert-module-globals` project.
               *
               * [1] https://github.com/substack/insert-module-globals/pull/31
               */
              process: function() {
                return JSON.stringify({env: {NODE_ENV: "production"}});
              }
            }
          },
          transform: [ require("grunt-react").browserify ],
          watch: true
        }
      },
      test: {
        src: [ "<%= conf.build %>/suite.js"],
        dest: "<%= conf.pub %>test/suite.js",
        options: {
          standalone: "tests",
          transform: [ require("grunt-react").browserify ]
        }
      }
    },
    uglify: {
      app: {
        files: {
          "<%= conf.pub %><%= pkg.name %>.min.js": "<%= conf.compiled %>"
        }
      }
    },
    // This is required to pre-convert JSX to allow jshint to consume correctly.
    react: {
      lint: {
        files: [
          {
            expand: true,
            cwd: "<%= conf.src %>",
            src: ["**/*.jsx"],
            dest: "<%= conf.build %>",
            ext: ".js"
          }
        ]
      }
    },
    jshint: {
      app: [
        "Gruntfile.js",
        "<%= conf.build %>**/*.js", // JSX files
        "<%= conf.src %>/**/*.js" // Exclude the JSX
      ],
      tests: [ "<%= conf.tests %>" ],
      options: {
        jshintrc: ".jshintrc"
      }
    },
    "generate-suite": {
      tests: {
        src: [ "<%= conf.tests %>" ],
        dest: "<%= conf.build %>/suite.js"
      }
    },
    mocha: {
      test: {
        options: {
          urls: [ "http://localhost:9610" ]
        }
      }
    }
    //,
    // connect: {
    //   server: {
    //     options: {
    //       port: 8000,
    //       base: "<%= conf.pub %>"
    //     }
    //   }
    // }
  });

  grunt.registerTask("buildtests", [
    "jshint:tests",
    "generate-suite",
    "browserify:test"
  ]);

  grunt.registerTask("runtests", [
    "buildtests",
    "mocha"
  ]);

  grunt.registerTask("dev", [
    "browserify:app",
    "less",
    "copy:css",
    "copy:html",
    "copy:fonts",
    "copy:img"
  ]);

  grunt.registerTask("run", [
    "dev",
    // "connect:server",
    "watch"
  ]);

  grunt.registerTask("dist", [
    "dev",
    "jshint",
    "uglify"
  ]);

  grunt.registerTask("check", []);

};
