$(function () {

    // 如果是IE浏览器，则不允许使用
    //var ieVersions = new Array("6.0", "7.0", "8.0");
    if ($.browser.msie) { //&& $.inArray($.browser.version, ieVersions)) {
        alertify.set({ labels: {
            "ok": "知道了"
        } });
        alertify.alert("请使用Google Chrome或Mozilla Firefox浏览器！暂不支持IE。");
        return false;
    }

    // 显示提示
    function displayTips(tips) {
        $('#result').remove();
        $('.ui-box').after($('<div></div>', {
            'class': 'alert alert-error',
            'id': 'result',
            'text': tips
        }));
    }

    // 预览/预览修改
    $('#preview, #edit-preview').on('click', function (e) {
        e.preventDefault();
        var thisId = $(this).attr('id');
        $('#result').remove();

        var servers = $.trim($('#server-list').val());
        if (servers === '') {
            displayTips('输入不能为空！');
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
            if (thisId === 'preview') {
                $('#preview').hide();
                $('#submit').show();
                $('#step-back').show();
            } else {
                $('#edit-preview').hide();
                $('#edit-cancel').hide();
                $('#edit-submit').show();
                $('#edit-step-back').show();
            }

            var tbody = $('<tbody></tbody>');
            var serverArray = new Array();
            for (var index in ipPortArray) {
                var ipPort = ipPortArray[index];
                var tr = $('<tr></tr>');
                tr.append('<td>' + (parseInt(index) + 1) + '</td>' + '<td>' + ipPort[0] + '</td><td>' + ipPort[1] + '</td>');
                tbody.append(tr);
                serverArray.push(ipPort.join(':'));
            }
            $('#preview-table > table > tbody').remove();
            $('#preview-table > table').append(tbody);

            $('#serversToSubmit').remove();
            $('#preview-table > table').after($('<input />', {
                'type': 'hidden',
                'id': 'serversToSubmit',
                'value': serverArray.join('-')
            }));
            $('#preview-table').slideDown();
        }
    });

    // 回退/修改回退 到预览功能
    $('#step-back, #edit-step-back').on('click', function(e){
        e.preventDefault();
        var thisId = $(this).attr("id");

        $('#preview-table').slideUp();
        if(thisId === 'step-back'){
            $('#submit').hide();
            $('#step-back').hide();
            $('#preview').show();
        }else{
            $('#edit-submit').hide();
            $('#edit-step-back').hide();
            $('#edit-preview').show();
            $('#edit-cancel').show();
        }
    });

    // 提交/修改提交
    $('#submit, #edit-submit').on('click', function (e) {
        e.preventDefault();
        var thisId = $(this).attr("id");

        var servers = $.trim($('#serversToSubmit').val()),
            comment = $.trim($('#comment').val()),
            logOrNot = $("input[name='lonOptionRadios']:checked").val(),
            assignMethod = $("input[name='amOptionsRadios']:checked").val();
        var specPort = "-1";
        var businessType = "-1";
        // 如果是指定端口方式
        if (assignMethod === "0") {
            specPort = $.trim($("input[name='port']").val());
            if (specPort === "") {
                alertify.log('请指定端口!', 'error', 3000);
                return false;
            }
            if (! /^\d{4,5}$/.test(specPort)){
                alertify.log('指定的端口应在1000-99999之间（包含1000和99999）', 'error', 3000);
                return false;
            }
        }
        // 如果是自动分配端口方式
        if (assignMethod === "1") {
            // 可能有业务区分，尝试获取业务类型
            businessType = $("#business-type > option:selected").val();
            if (businessType === undefined) {
                businessType = "";
            }
        }

        if (thisId === 'submit') {
            var req = $.ajax({
                'type': 'post',
                'url': '/applyvport',
                'data': {
                    autoornot: assignMethod,
                    business: businessType,
                    port: specPort,
                    servers: servers,
                    comment: comment,
                    logornot: logOrNot
                },
                'dataType': 'json'
            });

            req.done(function (resp) {
                var resultClass = 'alert alert-error';
                // 如果成功，则清空原表单数据
                if (resp.Success == 'true') {
                    $('#server-list').val('');
                    $('#comment').val('');
                    $('input[name="lonOptionRadios"][value="1"]').attr('checked', true);
                    resultClass = 'alert alert-success';
                }
                // 显示结果
                $('#result').remove();
                $('.ui-box').after($('<div></div>', {
                    'class': resultClass,
                    'id': 'result',
                    'text': resp.Msg
                }));

                $('#preview-table').slideUp();
                $('#submit').hide();
                $('#preview').show();
            });
        } else {
            var id = $("#idToEdit").val();
            var req = $.ajax({
                'type': 'post',
                'url': '/edittask',
                'data': {
                    servers: servers,
                    comment: comment,
                    logornot: logOrNot,
                    id: id
                },
                'dataType': 'json'
            });
            req.done(function (resp) {
                $('#preview-table').slideUp();
                $('#edit-submit').hide();
                $('#edit-preview').show();
                if (resp.Success === 'true') {
                    alertify.log(resp.Msg, 'success', 1000);
                    setTimeout("window.location.href='/listenlist'", 1500);
                } else {
                    alertify.log(resp.Msg, 'error', 5000);
                }
            });
        }
    });

    // 点击触发编辑/删除功能
    $('.date-time').popover({
        html: true,
        placement: 'right',
        trigger: 'manual',
        content: '<input type="button" class="btn btn-danger" id="del-this-task" value="删除" /> <input type="button" class="btn btn-warning" id="edit-this-task" value="编辑" />'
    }).click(function (e) {
            $('.popover').remove();
            $(this).popover('toggle');
            e.preventDefault();
            e.stopPropagation();
        });
    //}

    // 删除任务
    $(document).on('click', '#del-this-task', function (e) {
        var idTd = $('.popover').siblings('.id'),
            id = $.trim(idTd.text());

        $('.popover').slideUp(200, function (e) {
            $('.popover').remove();
        });

        alertify.set({
            buttonReverse: true,
            labels: {
                ok: '是',
                cancel: '否'
            } });
        alertify.confirm('你确定删除该任务吗？', function (e) {
            if (e) {
                alertify.log('你选择了"是"', '', 2000);

                var req = $.ajax({
                    'url': '/dellistentask?taskid=' + id,
                    'dataType': 'json'
                });

                req.done(function (resp) {
                    if (resp.Success === 'true') {
                        idTd.parents('tr').remove();
                        alertify.log(resp.Msg, 'success', 3000);
                    } else {
                        alertify.log(resp.Msg, 'error', 5000);
                    }

                });

            } else {
                alertify.log('你选择了"否"', '', 2000);
            }
        });
    });

    // 编辑（修改）任务
    $(document).on("click", "#edit-this-task", function (e) {
        e.preventDefault();

        var parent = $(this).parents(".popover");
        var servers = parent.siblings(".servers").html().replace(/(<(br|BR)\s*\/?>)/g, '\n');

        var comment = parent.siblings(".comment").text();
        var logON = $.trim(parent.siblings(".logornot").text()),
            logOrNot = logON === '是' ? '1' : '0';

        $("#idToEdit").remove();
        $("#preview-table").append($("<input />", {
            'type': 'hidden',
            'id': 'idToEdit',
            'value': parent.siblings('.id').text()
        }));

        $("#server-list").val(servers);
        $("#comment").val(comment);
        $("input[name='lonOptionRadios'][value='" + logOrNot + "']").attr("checked", true);

        $("#listenlist-div").slideUp();
        $("#edit-div").slideDown();
    });

    // 取消编辑
    $('#edit-cancel').on('click', function(e){
        e.preventDefault();
        $("#edit-div").slideUp();
        $("#listenlist-div").slideDown();
    });

    // 生成最新配置应用到主haproxy或从haproxy
    $("#apply-to-master, #apply-to-slave").on("click", function (e) {
        if ($(this).attr("id") === 'apply-to-master') {
            var target = "master";
        } else {
            target = "slave";
        }
        var req = $.ajax({
            'url': '/applyconf?target=' + target,
            'dataType': 'json'
        });

        req.done(function (resp) {
            if (resp.Success === 'true') {
                alertify.log(resp.Msg, 'success', 3000);
            } else {
                alertify.log(resp.Msg, 'error', 5000);
            }
        });
    });

    // 选择自动分配端口或指定端口
    $("input[name='amOptionsRadios']").on("change", function (e) {
        var assignMethod = $("input[name='amOptionsRadios']:checked").val();
        if (assignMethod === "0") {
            $("#to-specify-port").show();
            $("#business-block").hide();
        } else {
            $("#to-specify-port").hide();
            $("#business-block").show();
        }
    });
});
