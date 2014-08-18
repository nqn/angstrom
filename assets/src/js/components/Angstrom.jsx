/** @jsx React.DOM */

var React = require("react");
var _ = require("underscore");
var Chart = require("./Chart.jsx");
var DataSeries = require("./DataSeries.jsx");
var BackboneMixin = require("../mixins/BackboneMixin");
var DataCenterMetricsCollection = require("../models/DataCenterMetricsCollection");
var d3 = require("d3");

module.exports = React.createClass({
  displayName: "Angstrom",

  getDefaultProps: function() {
    return {
      interpolate: "linear",
      width: 600,
      height: 300
    };
  },

  render: function() {


    // <Chart width={this.props.width} height={this.props.height}>
    //   <DataSeries data={data.get("TotalCpus").get("series")} size={size} xScale={xScale} yScale={yScale} ref="series1" color="red" fill="red" />
    // </Chart>

    var data = new DataCenterMetricsCollection();
    data.fetch();
    console.log(data.get("TotalMemory").get("series"), data.get("TotalDisk").get("series"));//data.get("TotalCpus").get("series"),

      var size = { width: this.props.width, height: this.props.height };

      var max = _.chain(data.get("TotalMemory").get("series"), data.get("TotalDisk").get("series"))
        .zip()
        .map(function(values) {
          return _.reduce(values, function(memo, value) { return Math.max(memo, value.y); }, 0);
        })
        .max()
        .value();

      var xScale = d3.scale.linear()
        .domain([0, data.get("TotalDisk").get("series").length])
        .range([0, this.props.width]);

      var yScale = d3.scale.linear()
        .domain([0, max])
        .range([0, this.props.height]);

    /* jshint trailing:false, quotmark:false, newcap:false */
    return (
      <div>
        <nav className="navbar navbar-inverse" role="navigation">
         <div className="container-fluid">
            <a className="navbar-brand" href="/">
              Ångström
            </a>
          </div>
        </nav>
        <div className="container-fluid">
          <Chart width={this.props.width} height={this.props.height}>
            <DataSeries data={data.get("TotalMemory").get("series")} size={size} xScale={xScale} yScale={yScale} ref="series2" color="green" fill="green" />
            <DataSeries data={data.get("TotalDisk").get("series")} size={size} xScale={xScale} yScale={yScale} ref="series3" color="cornflowerblue" fill="lightsteelblue" />
          </Chart>
        </div>
      </div>
    );
  }
});
