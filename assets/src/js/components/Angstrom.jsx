/** @jsx React.DOM */

var React = require("react");
var _ = require("underscore");
var Chart = require("./Chart.jsx");
var BackboneMixin = require("../mixins/BackboneMixin");
var DataCenterMetricsCollection = require("../models/DataCenterMetricsCollection");

var UPDATE_INTERVAL = 5000;

module.exports = React.createClass({
  displayName: "Angstrom",

  mixins: [BackboneMixin],

  getInitialState: function() {
    return {
      collection: new DataCenterMetricsCollection()
    };
  },

  getBackboneModels: function () {
    return this.state.collection;
  },

  fetchResource: function () {
    this.state.collection.fetch({
      reset: true,
      success: function () {
        this._boundForceUpdate();
      }.bind(this)
    });
  },

  componentDidMount: function() {
      this.startPolling();
  },

  componentWillUnmount: function() {
    this.stopPolling();
  },

  startPolling: function() {
    if (this._interval == null) {
      this.fetchResource();
      this._interval = setInterval(this.fetchResource, UPDATE_INTERVAL);
    }
  },

  stopPolling: function() {
    if (this._interval != null) {
      clearInterval(this._interval);
      this._interval = null;
    }
  },

  render: function() {

    var collection = this.state.collection.models;
    var revisedData = [];

    var CpusDef = [];
    var MemoryDef = [];
    var data = [];


    if (collection.length > 0) {
      CpusDef = [
        {
          name: "SlackCpus",
          color: "#d9534f",
          // renderer: "value"
        },
        {
          name: "UsedCpus",
          color: "#faef66",
          // renderer: "value"
        },
        {
          name: "AllocatedCpus",
          color: "#00b482",
          // renderer: "value"
        },
        {
          name: "TotalCpus",
          color: "#428bca",
          // stroke: 'rgba(0,0,0,0.15)',
          // renderer: "value"
        }
      ];

      MemoryDef = [
        {
          name: "SlackMemory",
          color: "#d9534f",
          // renderer: "value"
        },
        {
          name: "UsedMemory",
          color: "#faef66",
          // renderer: "value"
        },
        {
          name: "AllocatedMemory",
          color: "#00b482",
          // renderer: "value"
        },
        {
          name: "TotalMemory",
          color: "#428bca",
          // renderer: "value"
        }
      ];

      // extract attributes
      _.each(collection, function (child) {
        data.push(child.attributes);
      });
    }

    /* jshint trailing:false, quotmark:false, newcap:false */
    return (
      <div>
        <nav className="navbar navbar-inverse" role="navigation">
         <div className="container">
            <a className="navbar-brand media" href="/">
              <div className="pull-left">
                <img src="img/angstrom-2.png" className="media-object" alt="Angstrom logo" height="25" width="28" />
              </div>
              <div className="media-body">
                Ångström
              </div>
            </a>
          </div>
        </nav>
        <div className="container">
          {
            collection.length > 0 ?
            <div>
              <Chart definitions={CpusDef} collection={data} />
              <Chart definitions={MemoryDef} collection={data} />
            </div> :
              null
          }
        </div>
      </div>
    );
  }
});
