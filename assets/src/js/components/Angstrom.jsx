/** @jsx React.DOM */

var React = require("react");
var _ = require("underscore");
var Chart = require("./Chart.jsx");
var DataSeries = require("./DataSeries.jsx");
var BackboneMixin = require("../mixins/BackboneMixin");
var d3 = require("d3");
require("../libs/nv.d3");
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

  getDefaultProps: function () {
    return {
      interpolate: "linear",
      width: 600,
      height: 300
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
    if (collection.models.length > 0) {
      var data = [
        {
          key: "TotalCpus",
          values: collection.get("TotalCpus").get("series")
        },
        {
          key: "AllocatedCpus",
          values: collection.get("AllocatedCpus").get("series")
        },
        {
          key: "UsedCpus",
          values: collection.get("UsedCpus").get("series")
        }
      ];
      nv.addGraph(function() {
        var chart = nv.models.stackedAreaChart()
                      .margin({right: 100})
                      .x(function(d) { return d.x; })   //We can modify the data accessor functions...
                      .y(function(d) { return d.y; })   //...in case your data is formatted differently.
                      .useInteractiveGuideline(true)    //Tooltips which show all data points. Very nice!
                      .rightAlignYAxis(true)      //Let's move the y-axis to the right side.
                      .transitionDuration(500)
                      .showControls(true)       //Allow user to choose 'Stacked', 'Stream', 'Expanded' mode.
                      .clipEdge(true);

        //Format x-axis labels with custom function.
        chart.xAxis
            .tickFormat(d3.format(",.2f"));

        chart.yAxis
            .tickFormat(d3.format(",.2f"));

        d3.select("#chart svg")
          .datum(data)
          .call(chart);
      });
    }

    /* jshint trailing:false, quotmark:false, newcap:false */
    return (
      <div>
        <nav className="navbar navbar-inverse" role="navigation">
         <div className="container-fluid">
            <img src="img/angstrom-2.png" />
            <a className="navbar-brand" href="/">
              Ångström
            </a>
          </div>
        </nav>
        <div className="container-fluid">
          {
            collection.models.length > 0 ?
              <div id="chart">
                <svg />
              </div> :
              null
          }
        </div>
      </div>
    );
  }
});
