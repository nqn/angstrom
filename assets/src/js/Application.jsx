/** @jsx React.DOM */

var $ = require("jquery");
var _ = require("underscore");
var Backbone = require("backbone");
var React = require("react");

Backbone.$ = $;

var Angstrom = require("./components/Angstrom.jsx");

var ApplicationRouter = Backbone.Router.extend({
  routes: {
    "" : "home"
  },

  initialize: function(options) {
    this.options = options;
  },

  home: function() {
    /* jshint trailing:false, quotmark:false, newcap:false */
    React.renderComponent(
      <Angstrom />,
      document.getElementById("angstrom")
    );
  }

});

this.router = new ApplicationRouter();
Backbone.history.start({ pushState: true });
