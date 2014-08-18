/** @jsx React.DOM */

var React = require("react");

module.exports = React.createClass({
  getDefaultProps: function() {
    return {
      path: "",
      color: "blue",
      width: 2,
      fill: "none"
    };
  },

  render: function() {
    return (
      <path d={this.props.path} stroke={this.props.color} strokeWidth={this.props.width} fill={this.props.fill} />
    );
  }
});
