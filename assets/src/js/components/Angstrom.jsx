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

    var collection = this.state.collection;
    var Cpus = [];
    var Memory = [];
    if (collection.models.length > 0) {
      Cpus = [
        {
          key: "TotalCpus",
          values: collection.get("TotalCpus").get("series"),
          color: "#428bca"
        },
        {
          key: "AllocatedCpus",
          values: collection.get("AllocatedCpus").get("series"),
          color: "#00b482"
        },
        {
          key: "UsedCpus",
          values: collection.get("UsedCpus").get("series"),
          color: "#d9534f"
        }
      ];

      Memory = [
        {
          key: "TotalMemory",
          values: collection.get("TotalMemory").get("series"),
          color: "#428bca"
        },
        {
          key: "AllocatedMemory",
          values: collection.get("AllocatedMemory").get("series"),
          color: "#00b482"
        },
        {
          key: "UsedMemory",
          values: collection.get("UsedMemory").get("series"),
          color: "#d9534f"
        }
      ];
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
            collection.models.length > 0 ?
            <div>
              <Chart collection={Cpus} />
              <Chart collection={Memory} />
            </div> :
              null
          }
        </div>
      </div>
    );
  }
});
