$(function () {
    // 显示提示
    function displayTips(tips) {
        $('#result').remove();
        $('.ui-box').after($('<div></div>', {
            'class': 'alert alert-error',
            'id': 'result',
            'text': tips
        }));
    }

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
            } else {
                $('#edit-preview').hide();
                $('#edit-submit').show();
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
            $('tbody').remove();
            $('table').append(tbody);

            $('#serversToSubmit').remove();
            $('table').after($('<input />', {
                'type': 'hidden',
                'id': 'serversToSubmit',
                'value': serverArray.join('-')
            }));
            $('#preview-table').slideDown();
        }
    });

    $('#submit, #edit-submit').on('click', function (e) {
        e.preventDefault();
        var thisId = $(this).attr("id");

        var servers = $('#serversToSubmit').val(),
            comment = $.trim($('#comment').val()),
            logOrNot = $("input[name='optionRadios']:checked").val();
        if (thisId === 'submit') {
            var req = $.ajax({
                'type': 'post',
                'url': '/applyvport',
                'data': {
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

    $(document).on('click', '#del-this-task', function (e) {
        var vportTd = $('.popover').siblings('.vport'),
            vport = $.trim(vportTd.text());

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
                    url: '/dellistentask?taskvport=' + vport,
                    'dataType': 'json'
                });

                req.done(function (resp) {
                    if (resp.Success === 'true') {
                        vportTd.parents('tr').remove();
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

    $(document).on("click", "#edit-this-task", function (e) {
        e.preventDefault();

        var parent = $(this).parents(".popover");
        var servers = parent.siblings(".servers").html().replace(/(<(br|BR)\s*\/?>)/g, '\n');
        var comment = parent.siblings(".comment").html();
        var logON = $.trim(parent.siblings(".logornot").text()),
            logOrNot = logON === '是' ? '1' : '0';

        $("#idToEdit").remove();
        $("#preview-table").append($("<input />", {
            'type': 'hidden',
            'id': 'idToEdit',
            'value': parent.siblings('.id').text()
        }));

        console.log(servers);
        console.log(comment);
        console.log(logON);
        console.log(logOrNot);

        $("#server-list").val(servers);
        $("#comment").val(comment);
        $("input[name='optionRadios'][value='" + logOrNot + "']").attr("checked", true);

        $("#listenlist-div").slideUp();
        $("#edit-div").slideDown();
    });

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

        req.done(function(resp){
            if (resp.Success === 'true') {
                alertify.log(resp.Msg, 'success', 3000);
            } else {
                alertify.log(resp.Msg, 'error', 5000);
            }
        });
    });
});
