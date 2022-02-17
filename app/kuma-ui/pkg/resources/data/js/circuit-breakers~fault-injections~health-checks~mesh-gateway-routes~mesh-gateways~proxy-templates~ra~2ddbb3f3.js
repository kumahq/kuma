(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["circuit-breakers~fault-injections~health-checks~mesh-gateway-routes~mesh-gateways~proxy-templates~ra~2ddbb3f3"],{"14eb":function(e,t,n){"use strict";var a=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("div",[n("LabelList",{attrs:{"has-error":e.hasError,"is-loading":e.isLoading,"is-empty":!e.hasDataplanes}},[n("ul",[n("li",[n("h4",[e._v("Dataplanes")]),n("input",{directives:[{name:"model",rawName:"v-model",value:e.searchInput,expression:"searchInput"}],staticClass:"k-input mb-4",attrs:{id:"dataplane-search",type:"text",placeholder:"Filter by name",required:""},domProps:{value:e.searchInput},on:{input:function(t){t.target.composing||(e.searchInput=t.target.value)}}}),e._l(e.filteredDataplanes,(function(t,a){return n("p",{key:a,staticClass:"my-1",attrs:{"data-testid":"dataplane-name"}},[n("router-link",{attrs:{to:{name:"dataplanes",query:{ns:t.dataplane.name},params:{mesh:t.dataplane.mesh}}}},[e._v(" "+e._s(t.dataplane.name)+" ")])],1)}))],2)])])],1)},r=[],o=(n("4de4"),n("caad"),n("b0c0"),n("2532"),n("96cf"),n("c964")),i=n("0f82"),s=n("0ada"),c={name:"PolicyConnections",components:{LabelList:s["a"]},props:{mesh:{type:String,required:!0},policyType:{type:String,required:!0},policyName:{type:String,required:!0}},data:function(){return{hasDataplanes:!1,isLoading:!0,hasError:!1,dataplanes:[],searchInput:""}},computed:{filteredDataplanes:function(){var e=this.searchInput.toLowerCase();return this.dataplanes.filter((function(t){var n=t.dataplane.name;return n.toLowerCase().includes(e)}))}},watch:{policyName:function(){this.fetchPolicyConntections()}},mounted:function(){this.fetchPolicyConntections()},methods:{fetchPolicyConntections:function(){var e=this;return Object(o["a"])(regeneratorRuntime.mark((function t(){var n,a,r;return regeneratorRuntime.wrap((function(t){while(1)switch(t.prev=t.next){case 0:return e.hasError=!1,e.isLoading=!0,t.prev=2,t.next=5,i["a"].getPolicyConnections({mesh:e.mesh,policyType:e.policyType,policyName:e.policyName});case 5:n=t.sent,a=n.items,r=n.total,e.hasDataplanes=r>0,e.dataplanes=a,t.next=15;break;case 12:t.prev=12,t.t0=t["catch"](2),e.hasError=!0;case 15:return t.prev=15,e.isLoading=!1,t.finish(15);case 18:case"end":return t.stop()}}),t,null,[[2,12,15,18]])})))()}}},l=c,u=n("2877"),p=Object(u["a"])(l,a,r,!1,null,null,null);t["a"]=p.exports},"23d6":function(e,t,n){"use strict";n("b91b")},"2f4c":function(e,t,n){"use strict";n("9863")},"536d":function(e,t,n){"use strict";n("66ad")},"62e5":function(e,t){var n;n=function(){function e(e,t){var n,a,r,o,i,s,c,l,u;null==t&&(t=""),r="",i=e.length,s=null,a=0,o=0;while(o<i){if(n=e.charAt(o),"\\"===n)r+=e.slice(o,+(o+1)+1||9e9),o++;else if("("===n)if(o<i-2)if(l=e.slice(o,+(o+2)+1||9e9),"(?:"===l)o+=2,r+=l;else if("(?<"===l){a++,o+=2,c="";while(o+1<i){if(u=e.charAt(o+1),">"===u){r+="(",o++,c.length>0&&(null==s&&(s={}),s[c]=a);break}c+=u,o++}}else r+=n,a++;else r+=n;else r+=n;o++}this.rawRegex=e,this.cleanedRegex=r,this.regex=new RegExp(this.cleanedRegex,"g"+t.replace("g","")),this.mapping=s}return e.prototype.regex=null,e.prototype.rawRegex=null,e.prototype.cleanedRegex=null,e.prototype.mapping=null,e.prototype.exec=function(e){var t,n,a,r;if(this.regex.lastIndex=0,n=this.regex.exec(e),null==n)return null;if(null!=this.mapping)for(a in r=this.mapping,r)t=r[a],n[a]=n[t];return n},e.prototype.test=function(e){return this.regex.lastIndex=0,this.regex.test(e)},e.prototype.replace=function(e,t){return this.regex.lastIndex=0,e.replace(this.regex,t)},e.prototype.replaceAll=function(e,t,n){var a;null==n&&(n=0),this.regex.lastIndex=0,a=0;while(this.regex.test(e)&&(0===n||a<n))this.regex.lastIndex=0,e=e.replace(this.regex,t),a++;return[e,a]},e}(),e.exports=n},6524:function(e,t,n){"use strict";var a=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("KButton",{staticClass:"docs-link",attrs:{appearance:"secondary",size:"small",target:"_blank",to:e.href},scopedSlots:e._u([{key:"icon",fn:function(){return[n("KIcon",{attrs:{"view-box":"0 0 14 14",icon:"externalLink"}})]},proxy:!0}])},[e._v(" Documentation ")])},r=[],o={name:"Documentation",props:{href:{type:String,required:!0}}},i=o,s=(n("2f4c"),n("2877")),c=Object(s["a"])(i,a,r,!1,null,"1920c6c9",null);t["a"]=c.exports},6663:function(e,t,n){"use strict";var a=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("div",{attrs:{"data-testid":"entity-url-control"}},[n("KClipboardProvider",{scopedSlots:e._u([{key:"default",fn:function(t){var a=t.copyToClipboard;return[n("KPop",{attrs:{placement:"bottom"},scopedSlots:e._u([{key:"content",fn:function(){return[n("div",[n("p",[e._v(e._s(e.confirmationText))])])]},proxy:!0}],null,!0)},[n("KButton",{attrs:{appearance:"secondary",size:"small"},on:{click:function(){a(e.shareUrl)}},scopedSlots:e._u([{key:"icon",fn:function(){return[n("KIcon",{attrs:{"view-box":"0 0 16 16",icon:"externalLink"}})]},proxy:!0}],null,!0)},[e._v(" "+e._s(e.copyButtonText)+" ")])],1)]}}])})],1)},r=[],o=(n("99af"),n("b0c0"),n("ac1f"),n("5319"),{name:"EntityURLControl",props:{name:{type:String,default:""},copyButtonText:{type:String,default:"Copy URL"},confirmationText:{type:String,default:"URL copied to clipboard!"},mesh:{type:String,default:""}},computed:{shareUrl:function(){var e="".concat(window.location.href.replace(window.location.hash,""),"#"),t=this.$router.resolve({name:this.$route.name,params:{mesh:this.mesh},query:{ns:this.name}}).resolved.fullPath;return"".concat(e).concat(t)}}}),i=o,s=n("2877"),c=Object(s["a"])(i,a,r,!1,null,null,null);t["a"]=c.exports},"66ad":function(e,t,n){},"6d8a":function(e,t,n){var a,r;r=n("62e5"),a=function(){var e;function t(){}return t.LIST_ESCAPEES=["\\","\\\\",'\\"','"',"\0","","","","","","","","\b","\t","\n","\v","\f","\r","","","","","","","","","","","","","","","","","","",(e=String.fromCharCode)(133),e(160),e(8232),e(8233)],t.LIST_ESCAPED=["\\\\",'\\"','\\"','\\"',"\\0","\\x01","\\x02","\\x03","\\x04","\\x05","\\x06","\\a","\\b","\\t","\\n","\\v","\\f","\\r","\\x0e","\\x0f","\\x10","\\x11","\\x12","\\x13","\\x14","\\x15","\\x16","\\x17","\\x18","\\x19","\\x1a","\\e","\\x1c","\\x1d","\\x1e","\\x1f","\\N","\\_","\\L","\\P"],t.MAPPING_ESCAPEES_TO_ESCAPED=function(){var e,n,a,r;for(a={},e=n=0,r=t.LIST_ESCAPEES.length;0<=r?n<r:n>r;e=0<=r?++n:--n)a[t.LIST_ESCAPEES[e]]=t.LIST_ESCAPED[e];return a}(),t.PATTERN_CHARACTERS_TO_ESCAPE=new r("[\\x00-\\x1f]|Â|Â |â¨|â©"),t.PATTERN_MAPPING_ESCAPEES=new r(t.LIST_ESCAPEES.join("|").split("\\").join("\\\\")),t.PATTERN_SINGLE_QUOTING=new r("[\\s'\":{}[\\],&*#?]|^[-?|<>=!%@`]"),t.requiresDoubleQuoting=function(e){return this.PATTERN_CHARACTERS_TO_ESCAPE.test(e)},t.escapeWithDoubleQuotes=function(e){var t;return t=this.PATTERN_MAPPING_ESCAPEES.replace(e,function(e){return function(t){return e.MAPPING_ESCAPEES_TO_ESCAPED[t]}}(this)),'"'+t+'"'},t.requiresSingleQuoting=function(e){return this.PATTERN_SINGLE_QUOTING.test(e)},t.escapeWithSingleQuotes=function(e){return"'"+e.replace(/'/g,"''")+"'"},t}(),e.exports=a},9863:function(e,t,n){},b91b:function(e,t,n){},e80b:function(e,t,n){var a=n("6d8a"),r="  ";function o(e){var t=typeof e;return e instanceof Array?"array":"string"==t?"string":"boolean"==t?"boolean":"number"==t?"number":"undefined"==t||null===e?"null":"hash"}function i(e,t){var n=o(e);switch(n){case"array":s(e,t);break;case"hash":c(e,t);break;case"string":u(e,t);break;case"null":t.push("null");break;case"number":t.push(e.toString());break;case"boolean":t.push(e?"true":"false");break}}function s(e,t){for(var n=0;n<e.length;n++){var a=e[n],o=[];i(a,o);for(var s=0;s<o.length;s++)t.push((0==s?"- ":r)+o[s])}}function c(e,t){for(var n in e){var a=[];if(e.hasOwnProperty(n)){var s=e[n];i(s,a);var c=o(s);if("string"==c||"null"==c||"number"==c||"boolean"==c)t.push(l(n)+": "+a[0]);else{t.push(l(n)+": ");for(var u=0;u<a.length;u++)t.push(r+a[u])}}}}function l(e){return e.match(/^[\w]+$/)?e:a.requiresDoubleQuoting(e)?a.escapeWithDoubleQuotes(e):a.requiresSingleQuoting(e)?a.escapeWithSingleQuotes(e):e}function u(e,t){t.push(l(e))}var p=function(e){"string"==typeof e&&(e=JSON.parse(e));var t=[];return i(e,t),t.join("\n")};e.exports=p},ff9d:function(e,t,n){"use strict";var a=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("div",{staticClass:"yaml-view"},[e.isReady?n("div",{staticClass:"yaml-view-content"},[e.isLoading||e.isEmpty?e._e():n("KCard",{attrs:{title:e.yamlTitle,"border-variant":"noBorder"},scopedSlots:e._u([{key:"body",fn:function(){return[n("KTabs",{key:e.environment,attrs:{tabs:e.tabs},scopedSlots:e._u([{key:"universal",fn:function(){return[n("KClipboardProvider",{scopedSlots:e._u([{key:"default",fn:function(t){var a=t.copyToClipboard;return[n("KPop",{attrs:{placement:"bottom"},scopedSlots:e._u([{key:"content",fn:function(){return[n("div",[n("p",[e._v("Entity copied to clipboard!")])])]},proxy:!0}],null,!0)},[n("KButton",{staticClass:"copy-button",attrs:{appearance:"primary",size:"small"},on:{click:function(){a(e.yamlContent.universal)}}},[e._v(" Copy Universal YAML ")])],1)]}}],null,!1,1536634960)}),n("Prism",{staticClass:"code-block",attrs:{language:"yaml",code:e.yamlContent.universal}})]},proxy:!0},{key:"kubernetes",fn:function(){return[n("KClipboardProvider",{scopedSlots:e._u([{key:"default",fn:function(t){var a=t.copyToClipboard;return[n("KPop",{attrs:{placement:"bottom"},scopedSlots:e._u([{key:"content",fn:function(){return[n("div",[n("p",[e._v("Entity copied to clipboard!")])])]},proxy:!0}],null,!0)},[n("KButton",{staticClass:"copy-button",attrs:{appearance:"primary",size:"small"},on:{click:function(){a(e.yamlContent.kubernetes)}}},[e._v(" Copy Kubernetes YAML ")])],1)]}}],null,!1,2265429040)}),n("Prism",{staticClass:"code-block",attrs:{language:"yaml",code:e.yamlContent.kubernetes}})]},proxy:!0}],null,!1,1506056494),model:{value:e.activeTab.hash,callback:function(t){e.$set(e.activeTab,"hash",t)},expression:"activeTab.hash"}})]},proxy:!0}],null,!1,137880475)})],1):e._e(),!0===e.loaders?n("div",[e.isLoading?n("KEmptyState",{attrs:{"cta-is-hidden":""},scopedSlots:e._u([{key:"title",fn:function(){return[n("div",{staticClass:"card-icon mb-3"},[n("KIcon",{attrs:{icon:"spinner",color:"rgba(0, 0, 0, 0.1)",size:"42"}})],1),e._v(" Data Loading... ")]},proxy:!0}],null,!1,3263214496)}):e._e(),e.isEmpty&&!e.isLoading?n("KEmptyState",{attrs:{"cta-is-hidden":""},scopedSlots:e._u([{key:"title",fn:function(){return[n("div",{staticClass:"card-icon mb-3"},[n("KIcon",{staticClass:"kong-icon--centered",attrs:{color:"var(--yellow-200)",icon:"warning","secondary-color":"var(--black-75)",size:"42"}})],1),e._v(" There is no data to display. ")]},proxy:!0}],null,!1,1612658095)}):e._e(),e.hasError?n("KEmptyState",{attrs:{"cta-is-hidden":""},scopedSlots:e._u([{key:"title",fn:function(){return[n("div",{staticClass:"card-icon mb-3"},[n("KIcon",{staticClass:"kong-icon--centered",attrs:{color:"var(--yellow-200)",icon:"warning","secondary-color":"var(--black-75)",size:"42"}})],1),e._v(" An error has occurred while trying to load this data. ")]},proxy:!0}],null,!1,822917942)}):e._e()],1):e._e()])},r=[],o=(n("caad"),n("a15b"),n("b0c0"),n("4fad"),n("ac1f"),n("2532"),n("1276"),n("f3f3")),i=n("2f62"),s=n("2ccf"),c=n.n(s),l=n("e80b"),u=n.n(l),p={name:"YamlView",components:{Prism:c.a},props:{title:{type:String,default:null},content:{type:Object,default:null},loaders:{type:Boolean,default:!0},isLoading:{type:Boolean,default:!1},hasError:{type:Boolean,default:!1},isEmpty:{type:Boolean,default:!1}},data:function(){return{tabs:[{hash:"#universal",title:"Universal"},{hash:"#kubernetes",title:"Kubernetes"}]}},computed:Object(o["a"])(Object(o["a"])({},Object(i["c"])({environment:"config/getEnvironment"})),{},{isReady:function(){return!this.isEmpty&&!this.hasError&&!this.isLoading},activeTab:{get:function(){var e=this.environment;return{hash:"#".concat(e),nohash:e}},set:function(e){return{hash:"#".concat(e),nohash:e}}},yamlTitle:function(){var e;return this.title?this.title:null!==(e=this.content)&&void 0!==e&&e.name?"Entity Overview for ".concat(this.content.name):"Entity Overview"},yamlContent:function(){var e=this,t=this.content,n=function(){var t={},n=Object.assign({},e.content),a=n.name,r=n.mesh,o=n.type,i=function(){var t=Object.assign({},e.content);return delete t.type,delete t.mesh,delete t.name,!!(t&&Object.entries(t).length>0)&&t};if(t.apiVersion="kuma.io/v1alpha1",t.kind=o,void 0!==r&&(t.mesh=n.mesh),null!==a&&void 0!==a&&a.includes(".")){var s=a.split("."),c=s.pop(),l=s.join(".");t.metadata={name:l,namespace:c}}else t.metadata={name:a};return i()&&(t.spec=i()),t},a={universal:u()(t),kubernetes:u()(n())};return a}})},d=p,f=(n("23d6"),n("536d"),n("2877")),h=Object(f["a"])(d,a,r,!1,null,"78c7b522",null);t["a"]=h.exports}}]);