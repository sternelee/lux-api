#!/bin/bash

# Lux API 部署脚本

set -e

echo "🚀 开始部署 Lux API 服务..."

# 检查是否安装了必要的工具
check_requirements() {
    echo "📋 检查部署要求..."
    
    # 检查 Go
    if ! command -v go &> /dev/null; then
        echo "❌ Go 未安装，请先安装 Go 1.24+"
        exit 1
    fi
    
    # 检查 SST
    if ! command -v sst &> /dev/null; then
        echo "❌ SST CLI 未安装，请先运行: npm install -g sst"
        exit 1
    fi
    
    # 检查 AWS CLI
    if ! command -v aws &> /dev/null; then
        echo "❌ AWS CLI 未安装，请先安装 AWS CLI"
        exit 1
    fi
    
    echo "✅ 所有要求都已满足"
}

# 构建项目
build_project() {
    echo "🔨 构建 Go 二进制文件..."
    
    # 设置环境变量
    export GOOS=linux
    export GOARCH=arm64
    export CGO_ENABLED=0
    
    # 清理旧的构建文件
    rm -f bootstrap
    
    # 构建
    go build -o bootstrap main.go
    
    if [ ! -f bootstrap ]; then
        echo "❌ 构建失败"
        exit 1
    fi
    
    echo "✅ 构建完成"
}

# 部署到 AWS
deploy_to_aws() {
    echo "☁️  部署到 AWS..."
    
    # 检查 AWS 凭证
    if ! aws sts get-caller-identity &> /dev/null; then
        echo "❌ AWS 凭证未配置，请先运行: aws configure"
        exit 1
    fi
    
    # 部署
    sst deploy
    
    echo "✅ 部署完成"
}

# 显示部署信息
show_deployment_info() {
    echo ""
    echo "🎉 部署成功！"
    echo ""
    echo "📊 获取 API URL:"
    echo "   sst output"
    echo ""
    echo "🔍 查看日志:"
    echo "   sst logs"
    echo ""
    echo "🗑️  删除部署:"
    echo "   sst remove"
    echo ""
    echo "📖 使用说明请查看 API_README.md"
}

# 主函数
main() {
    check_requirements
    build_project
    deploy_to_aws
    show_deployment_info
}

# 运行主函数
main
