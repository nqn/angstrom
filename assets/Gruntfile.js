module.exports = function(grunt) {

  [ "grunt-contrib-watch",
    "grunt-browserify",
    "grunt-react",
    "grunt-contrib-uglify",
    "grunt-contrib-jshint",
    "grunt-contrib-copy",
    "grunt-contrib-compass",
    "grunt-contrib-connect",
    "grunt-mocha"
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
      sass: "<%= conf.src %>scss",
      css: "<%= conf.pub %>css",
      src: "src/",
      js: "<%= conf.src %>js/",
      pub: "public/",
      build: "build/"
    },
    compass: {
      dist: {
        options: {
          sassDir: "<%= conf.sass %>",
          cssDir: "<%= conf.css %>",
          outputStyle: "compressed"
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
      }
    },
    watch: {
      app: {
        files: [
          "<%= conf.sys %>",
          "<%= conf.compiled %>"
        ]
      },
      css: {
        files: ["**/*.sass", "**/*.scss"],
        tasks: ["compass"]
      },
      html: {
        files: ["<%= conf.src %>**/*.html"],
        tasks: ["copy:html"]
      },
      img: {
        files: ["<%= conf.src %>**/*"],
        tasks: ["copy:img"]
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
        "<%= conf.src %>/**/*.jsx" // Exclude the JSX
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
    "browserify:app"
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
