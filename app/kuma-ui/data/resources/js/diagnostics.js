(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["diagnostics"],{"1d10":function(t,n,e){"use strict";var a=function(){var t=this,n=t.$createElement,e=t._self._c||n;return e("div",{staticClass:"component-frame"},[t._t("default")],2)},o=[],i={name:"FrameSkeleton"},r=i,c=(e("8463"),e("2877")),s=Object(c["a"])(r,a,o,!1,null,"664e217a",null);n["a"]=s.exports},"38ba":function(t,n,e){},"43c3":function(t,n,e){"use strict";var a=function(){var t=this,n=t.$createElement,e=t._self._c||n;return e("header",{staticClass:"page-header",class:{"flex justify-between items-center my-6":!t.noflex}},[t._t("default")],2)},o=[],i={props:{noflex:{type:Boolean,default:!1}}},r=i,c=(e("e234"),e("2877")),s=Object(c["a"])(r,a,o,!1,null,null,null);n["a"]=s.exports},7927:function(t,n,e){},7991:function(t,n,e){"use strict";var a=e("ba38"),o=e.n(a);o.a},8463:function(t,n,e){"use strict";var a=e("7927"),o=e.n(a);o.a},"94b6":function(t,n,e){"use strict";e.r(n);var a=function(){var t=this,n=t.$createElement,e=t._self._c||n;return e("div",{staticClass:"local-cps"},[e("page-header",{attrs:{noflex:""}},[e("h2",{staticClass:"type-xxl"},[t._v(" "+t._s(t.pageTitle)+" ")])]),e("FrameSkeleton",{staticClass:"py-2 px-4"},[t.isReady?e("KCard",{attrs:{"border-variant":"noBorder"}},[e("template",{slot:"body"},[e("prism",{staticClass:"code-block",attrs:{language:"json",code:t.codeOutput}})],1),e("template",{slot:"actions"},[t.codeOutput?e("KClipboardProvider",{scopedSlots:t._u([{key:"default",fn:function(n){var a=n.copyToClipboard;return[e("KPop",{attrs:{placement:"bottom"}},[e("KButton",{attrs:{appearance:"primary"},on:{click:function(){a(t.codeOutput)}}},[t._v(" Copy config to clipboard ")]),e("div",{attrs:{slot:"content"},slot:"content"},[e("p",[t._v("Config copied to clipboard!")])])],1)]}}],null,!1,246741307)}):t._e()],1)],2):t._e(),t.isLoading||t.hasError?e("KEmptyState",{attrs:{"cta-is-hidden":""}},[e("template",{slot:"title"},[e("div",{staticClass:"card-icon mb-3"},[t.icon?e("KIcon",{staticClass:"kong-icon--centered",attrs:{color:t.iconColor,icon:t.icon,size:"42"}}):t._e()],1),t.isLoading?e("span",[t._v(" Data Loading... ")]):t.hasError?e("span",[t._v(" An error has occurred while trying to load this data. ")]):t._e()])],2):t._e()],1)],1)},o=[],i=(e("d3b7"),e("f3f3")),r=e("2f62"),c=e("2ccf"),s=e.n(c),l=(e("a878"),e("43c3")),u=e("1d10"),f={name:"Diagnostics",metaInfo:{title:"Diagnostics"},components:{PageHeader:l["a"],FrameSkeleton:u["a"],Prism:s.a},data:function(){return{isLoading:!0,hasError:!1}},computed:Object(i["a"])(Object(i["a"])({},Object(r["b"])({config:"getConfig"})),{},{icon:function(){return this.isLoading?"spinner":!!this.hasError&&"warning"},iconColor:function(){return this.hasError?"var(--yellow-200)":"#ccc"},isReady:function(){return!this.hasError&&!this.isLoading},pageTitle:function(){var t=this.$route.meta.title;return t},codeOutput:function(){var t=this.config;return JSON.stringify(t,null,2)},configUrl:function(){var t=localStorage.getItem("kumaApiUrl")||null,n=t?"".concat(t,"/config"):null;return n}}),beforeMount:function(){this.fetchData()},methods:{fetchData:function(){var t=this;this.config?setTimeout((function(){t.isLoading=!1}),"500"):this.$store.dispatch("getConfig").catch((function(n){t.hasError=!0,console.log(n)})).finally((function(){setTimeout((function(){t.isLoading=!1}),"500")}))}}},d=f,p=(e("7991"),e("2877")),g=Object(p["a"])(d,a,o,!1,null,"de4cea76",null);n["default"]=g.exports},ba38:function(t,n,e){},e234:function(t,n,e){"use strict";var a=e("38ba"),o=e.n(a);o.a}}]);