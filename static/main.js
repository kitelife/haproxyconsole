$(function () {
    // 显示提示
    function displayTips(tips) {
        $("#result").remove();
        $(".ui-box").after($('<div></div>', {
            'class': "alert alert-error",
            'id': 'result',
            'text': tips
        }));
    }

    $("#preview").on("click", function (e) {
        e.preventDefault();

        $("#result").remove();

        var servers = $.trim($("#server-list").val());
        if (servers === '') {
            displayTips("输入不能为空！");
            return false;
        }

        // 注意正则表达式不要用引号包围
        var serverList = servers.split(/\r?\n/);
        var syntaxError = 0;
        var ipPortArray = new Array();
        for (var index in serverList) {
            var server = $.trim(serverList[index]);
            if (server.indexOf(':') === -1) {
                syntaxError = 1;
                break;
            }
            var ipPort = server.split(':');
            ipPort[0] = $.trim(ipPort[0]);
            ipPort[1] = $.trim(ipPort[1]);
            if (ipPort[0] === '' || ipPort[1] === '') {
                syntaxError = 1;
                break;
            }
            ipPortArray.push(ipPort);
        }
        if (syntaxError === 1) {
            displayTips('单行格式不对，应为ip:port');
            return false;
        } else {
            $("#preview").hide();
            $("#submit").show();

            var tbody = $("<tbody></tbody>");
            var serverArray = new Array();
            for (var index in ipPortArray) {
                var ipPort = ipPortArray[index];
                var tr = $("<tr></tr>");
                tr.append("<td>" + (parseInt(index) + 1) + "</td>" + "<td>" + ipPort[0] + "</td><td>" + ipPort[1] + "</td>");
                tbody.append(tr);
                serverArray.push(ipPort.join(':'));
            }
            $("tbody").remove();
            $("table").append(tbody);

            $("#serversToSubmit").remove();
            $("table").after($("<input />", {
                "type": "hidden",
                "id": "serversToSubmit",
                "value": serverArray.join('-')
            }));
            $("#preview-table").slideDown();
        }
    });

    $("#submit").on("click", function(e){
        e.preventDefault();

        var servers = $("#serversToSubmit").val();
        var req = $.ajax({
            'url': '/applyvport?servers=' + servers,
            'dataType': 'json'
        });

        req.done(function(resp){
            var resultClass = 'alert alert-error';
            // 如果成功，则清空原表单数据
            if (resp.Success == "true") {
                $("#server-list").val('');
                resultClass = 'alert alert-success'
            }
            // 显示结果
            $("#result").remove();
            $(".content").append($('<div></div>', {
                'class': resultClass,
                'id': 'result',
                'text': resp.Msg
            }));

            $("#preview-table").slideUp();
            $("#submit").hide();
            $("#preview").show();
        });
    });
    
    $(".date-time").popover({
        html: true,
        placement: 'right',
        trigger: 'hover',
        content: '<input type="button" class="btn btn-danger" id="del-this-task" value="删除该任务" />'
    });
});