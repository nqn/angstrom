/** @jsx React.DOM */

var React = require("react");
var d3 = require("d3");
require("../libs/nv.d3");

module.exports = React.createClass({
  displayName: "Chart",

  propTypes: {
    collection: React.PropTypes.object.isRequired,
    onSelectApp: React.PropTypes.func.isRequired
  },

  getInitialState: function() {
    return {
      chart: null
    };
  },

  getDefaultProps: function () {
    return {
      collection: [],
      width: 1200,
      height: 400
    };
  },

  componentDidMount: function() {
    nv.addGraph(function() {
      var chart = nv.models.stackedAreaChart()
                    .margin({right: 100})
                    .x(function(d) { return d.x; })   // We can modify the data accessor functions...
                    .y(function(d) { return d.y; })   // ...in case your data is formatted differently.
                    .useInteractiveGuideline(true)    // Tooltips which show all data points. Very nice!
                    .transitionDuration(500)
                    .showControls(false);       // Allow user to choose "Stacked", "Stream", "Expanded" mode.
                    // .clipEdge(true);

       // Format x-axis labels with custom function.
      chart.xAxis
        .tickFormat(function(d) {
          return d3.time.format("%I:%M:%S%p")(new Date(d)).replace(/^0+/, "");
      });

      chart.width( this.props.width );
      //   .height( this.props.height );

      chart.yAxis
          .tickFormat(d3.format(","));


      // set state
      this.setState({chart: chart});

      // nv.utils.windowResize(this.updateChart(this.props));


      d3.select(this.getDOMNode())
        .call(chart)
        .call(this.updateChart(this.props));
    }.bind(this));
  },

  componentWillUnmount: function() {

  },

  shouldComponentUpdate: function(props) {
    d3.select(this.getDOMNode())
      .call(this.updateChart(props));

    // always skip React's render step
    return false;
  },

  // d3 chart function
  // a higher-order function to allowing
  // passing in the component properties/state
  updateChart: function(props) {
    return function(node) {
      // update data set
      node.datum(props.collection);
      if (this.state.chart != null) {
        this.state.chart.update();
      }
    }.bind(this);
  },

  render: function() {
    var style = {
      height: this.props.height
    };
    return (
      <svg style={style} />
    );
  }
});
