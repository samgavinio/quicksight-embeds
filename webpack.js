const path = require('path');
const webpack = require('webpack');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const OptimizeCSSAssetsPlugin = require('optimize-css-assets-webpack-plugin');
const CopyWebpackPlugin = require('copy-webpack-plugin');

module.exports = {
  entry: {
    client: './public/assets/js/index.js'
  },
  target: 'node',
  output: {
    path: path.resolve('./public/dist'),
    filename: 'js/[name].compiled.js'
  },
  module: {
   rules: [
      {
        test: /\.js$/,
        loader: 'babel-loader',
        exclude: /node_modules/
      },
      {
        enforce: 'pre',
        test: /\.js$/,
        exclude: /node_modules/,
        loader: 'eslint-loader'
      },
      {
        test: /\.css$/,
        use: [
          { loader: MiniCssExtractPlugin.loader },
          'css-loader'
        ]
      }
    ]
  },
  resolve: {
    extensions: ['.js']
  },
  plugins: [
    new MiniCssExtractPlugin({
      filename: 'css/client.min.css'
    }),
    new CopyWebpackPlugin([
      {
        from:'public/assets/image',
        to:'image'
      },
      {
        from:'public/assets/js/vendor',
        to:'js/vendor'
      } 
    ]), 
  ],
  optimization: {
    minimizer: [
      new OptimizeCSSAssetsPlugin({})
    ]
  }
};
