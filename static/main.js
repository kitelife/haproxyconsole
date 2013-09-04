$(function () {
    $("#submit").on("click", function (e) {
        e.preventDefault();

        var masterIp = $.trim($("input[name='master-ip']").val()),
            masterPort = $.trim($("input[name='master-port']").val());
        var backupIp = $.trim($("input[name='backup-ip']").val()),
            backupPort = $.trim($("input[name='backup-port']").val());

        if (masterIp !== '' && masterPort !== '' && backupIp !== '' && backupPort !== '') {
            var req = $.ajax({
                'url': '/applyvport',
                'type': 'post',
                'data': {
                    masterIp: masterIp,
                    masterPort: masterPort,
                    backupIp: backupIp,
                    backupPort: backupPort
                },
                'dataType': 'json'
            });

            req.done(function (resp) {

                var resultClass = 'alert alert-error';
                // 如果成功，则清空原表单数据
                if (resp.Success == "true") {
                    $("input[name='master-ip']").val('');
                    $("input[name='master-port']").val('');
                    $("input[name='backup-ip']").val('');
                    $("input[name='backup-port']").val('');

                    resultClass = 'alert alert-success'
                }
                // 显示结果
                $("#result").remove();
                $(".content").append($('<div></div>', {
                    'class': resultClass,
                    'id': 'result',
                    'text': resp.Msg
                }));
            });
        }
    });

    $.validator.addMethod("ip", function (value, element) {
        var ip = /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;
        return this.optional(element) || (ip.test(value) && (RegExp.$1 < 256 && RegExp.$2 < 256 && RegExp.$3 < 256 && RegExp.$4 < 256));
    }, "Ip地址格式错误");

    $("#form").validate({
        rules: {
            'master-port': {
                required: true,
                number: true
            },
            'backup-port': {
                required: true,
                number: true
            },
            'master-ip': {
                required: true,
                ip: true
            },
            'backup-ip': {
                required: true,
                ip: true
            }
        },
        messages: {
            'master-port': {
                required: "请输入主机端口",
                number: "端口应为数字"
            },
            'backup-port': {
                required: "请输入备机端口",
                number: "端口应为数字"
            },
            'master-ip': {
                required: "请输入主机ip"
            },
            'backup-ip': {
                required: "请输入备机ip"
            }
        }
    });
});