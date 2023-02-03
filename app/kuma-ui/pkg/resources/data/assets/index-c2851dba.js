import{C as B,c as Z,g as J,d as X,a as S,T as w,_ as ee}from"./production-8efaeab1.js";import{d as b,c as A,u as t,o as r,a as g,b as M,w as p,e as f,f as i,g as o,h as _,t as h,r as U,i as V,p as x,j as O,k as z,l as W,n as te,F as N,m as q,q as D,s as E,v as se,x as oe,y as Q,T as ne,z as ae}from"./runtime-dom.esm-bundler-fd3ecc5a.js";import{c as ie,a as ce,u as K,b as Y}from"./vue-router-67937a96.js";import{s as P,u as $}from"./store-ec4aec64.js";import{V as re,m as T,_ as le,P as I,T as G,a as ue,y as H,c as _e,$ as de,b as pe}from"./kongponents.es-7ead79da.js";import{k as F}from"./kumaApi-08f7fc23.js";import{u as me,a as fe}from"./index-e2f1942d.js";import{u as L,a as he}from"./index-be4d4b11.js";import{_ as k}from"./_plugin-vue_export-helper-c27b6911.js";import{P as ge}from"./constants-31fdaf55.js";import{d as ve}from"./datadogLogEvents-302eea7b.js";import{A as ye,a as be}from"./AccordionItem-2fdfb42d.js";import"./vuex.esm-bundler-4e6e06ec.js";import"./DoughnutChart-210a9e41.js";(function(){const e=document.createElement("link").relList;if(e&&e.supports&&e.supports("modulepreload"))return;for(const c of document.querySelectorAll('link[rel="modulepreload"]'))l(c);new MutationObserver(c=>{for(const n of c)if(n.type==="childList")for(const d of n.addedNodes)d.tagName==="LINK"&&d.rel==="modulepreload"&&l(d)}).observe(document,{childList:!0,subtree:!0});function s(c){const n={};return c.integrity&&(n.integrity=c.integrity),c.referrerpolicy&&(n.referrerPolicy=c.referrerpolicy),c.crossorigin==="use-credentials"?n.credentials="include":c.crossorigin==="anonymous"?n.credentials="omit":n.credentials="same-origin",n}function l(c){if(c.ep)return;c.ep=!0;const n=s(c);fetch(c.href,n)}})();function Ae(a,e="/"){const s=ie({history:ce(e),routes:a});return s.beforeEach(Me),s.beforeEach(ke),s.beforeEach($e),s}const Me=function(a,e,s){a.fullPath.startsWith("/#/")?s(a.fullPath.substring(2)):s()},ke=function(a,e,s){a.params.mesh&&a.params.mesh!==P.state.selectedMesh&&P.dispatch("updateSelectedMesh",a.params.mesh),s()},$e=function(a,e,s){const l=P.state.onboarding.isCompleted,c=a.meta.onboardingProcess,n=P.getters.shouldSuggestOnboarding;l&&c&&!n?s({name:"home"}):!l&&!c&&n?s({name:B.get("onboardingStep")??"onboarding-welcome"}):s()},Se=b({__name:"AppBreadcrumbs",setup(a){const e=K(),s=Y(),l=A(()=>{const c=new Map;for(const n of e.matched){if(n.name==="home"||n.meta.parent==="home")continue;if(n.meta.parent!==void 0){const u=s.resolve({name:n.meta.parent});u.name&&c.set(u.name,{to:u,key:u.name,title:u.meta.title,text:u.meta.title})}if((n.name===e.name||n.redirect===e.name)&&n.meta.breadcrumbExclude!==!0&&e.name){let u=e.meta.title;e.meta.breadcrumbTitleParam&&e.params[e.meta.breadcrumbTitleParam]&&(u=e.params[e.meta.breadcrumbTitleParam]),c.set(e.name,{to:e,key:e.name,title:u,text:u})}}return Array.from(c.values())});return(c,n)=>t(l).length>0?(r(),g(t(re),{key:0,items:t(l)},null,8,["items"])):M("",!0)}}),we=o("p",null,"Unable to reach the API",-1),Ne={key:0},Ue=b({__name:"AppErrorMessage",setup(a){return(e,s)=>(r(),g(t(le),{class:"global-api-status empty-state--wide-content empty-state--compact","cta-is-hidden":""},{title:p(()=>[f(t(T),{class:"mb-3",icon:"warning",color:"var(--black-75)","secondary-color":"var(--yellow-300)",size:"64"}),i(),we]),message:p(()=>[o("p",null,[i(`
        Please double check to make sure it is up and running `),t(F).baseUrl?(r(),_("span",Ne,[i(", and it is reachable at "),o("code",null,h(t(F).baseUrl),1)])):M("",!0)])]),_:1}))}}),Ie={key:0,"data-testid":"notification-amount",class:"notification-icon__amount"},Le=b({__name:"NotificationIcon",setup(a){const e=$(),s=A(()=>e.getters["notifications/amountOfActions"]);function l(){e.dispatch("notifications/openModal")}return(c,n)=>(r(),_("button",{class:"notification-icon cursor-pointer",type:"button",onClick:l},[f(t(T),{icon:"notificationBell",color:"var(--yellow-300)"}),i(),t(s)>0?(r(),_("span",Ie,h(t(s)),1)):M("",!0)]))}});const Re=k(Le,[["__scopeId","data-v-cadae07a"]]),Te={class:"upgrade-check"},Ce={class:"alert-content"},Ee=b({__name:"UpgradeCheck",setup(a){const e=L(),s=U(""),l=U(!1);c();async function c(){try{s.value=await F.getLatestVersion()}catch(n){l.value=!1,console.error(n)}finally{if(s.value!==""){const n=Z(s.value,e("KUMA_VERSION"));l.value=n===1}else{const d=new Date,u=new Date("2020-06-03 12:00:00"),v=new Date(u.getFullYear(),u.getMonth()+3,u.getDate());l.value=d.getTime()>=v.getTime()}}}return(n,d)=>(r(),_("div",Te,[l.value?(r(),g(t(G),{key:0,class:"upgrade-check-alert",appearance:"warning",size:"small"},{alertMessage:p(()=>[o("div",Ce,[o("div",null,h(t(e)("KUMA_PRODUCT_NAME"))+` update available
          `,1),i(),o("div",null,[f(t(I),{class:"warning-button",appearance:"primary",size:"small",to:t(e)("KUMA_INSTALL_URL")},{default:p(()=>[i(`
              Update
            `)]),_:1},8,["to"])])])]),_:1})):M("",!0)]))}});const Pe=k(Ee,[["__scopeId","data-v-03ce650c"]]),xe=a=>(x("data-v-c348c723"),a=a(),O(),a),Oe={class:"app-header"},Ke={class:"horizontal-list"},De={class:"upgrade-check-wrapper"},Be={key:0,class:"horizontal-list"},Fe={class:"app-status app-status--mobile"},Ve={class:"app-status app-status--desktop"},ze=["href"],qe=["href"],Ye=xe(()=>o("span",{class:"visually-hidden"},"Diagnostics",-1)),Ge=b({__name:"AppHeader",setup(a){const[e,s]=[me(),fe()],l=$(),c=L(),n=A(()=>l.getters["notifications/amountOfActions"]>0),d=A(()=>{const v=l.getters["config/getEnvironment"];return v?v.charAt(0).toUpperCase()+v.substring(1):"Universal"}),u=A(()=>l.getters["config/getMulticlusterStatus"]?"Multi-Zone":"Standalone");return(v,m)=>{const y=V("router-link");return r(),_("header",Oe,[o("div",Ke,[f(y,{to:{name:"home"}},{default:p(()=>[f(t(e))]),_:1}),i(),f(t(s),{class:"gh-star",href:"https://github.com/kumahq/kuma","aria-label":"Star kumahq/kuma on GitHub"},{default:p(()=>[i(`
        Star
      `)]),_:1}),i(),o("div",De,[f(Pe)])]),i(),t(l).state.config.status==="OK"?(r(),_("div",Be,[o("div",Fe,[f(t(ue),{width:"280"},{content:p(()=>[o("p",null,[i(h(t(l).state.config.tagline)+" ",1),o("b",null,h(t(l).state.config.version),1),i(" on "),o("b",null,h(t(d)),1),i(" ("+h(t(u))+`)
            `,1)])]),default:p(()=>[f(t(I),{appearance:"outline"},{default:p(()=>[i(`
            Info
          `)]),_:1}),i()]),_:1})]),i(),o("p",Ve,[i(h(t(l).state.config.tagline)+" ",1),o("b",null,h(t(l).state.config.version),1),i(" on "),o("b",null,h(t(d)),1),i(" ("+h(t(u))+`)
      `,1)]),i(),t(n)?(r(),g(Re,{key:0})):M("",!0),i(),f(t(_e),{class:"help-menu",icon:"help","button-appearance":"outline","kpop-attributes":{placement:"bottomEnd"}},{items:p(()=>[f(t(H),null,{default:p(()=>[o("a",{href:`${t(c)("KUMA_DOCS_URL")}/?${t(c)("KUMA_UTM_QUERY_PARAMS")}`,target:"_blank",rel:"noopener noreferrer"},`
              Documentation
            `,8,ze)]),_:1}),i(),f(t(H),null,{default:p(()=>[o("a",{href:t(c)("KUMA_FEEDBACK_URL"),target:"_blank",rel:"noopener noreferrer"},`
              Feedback
            `,8,qe)]),_:1})]),_:1}),i(),f(t(I),{to:{name:"diagnostics"},icon:"gearFilled","button-appearance":"btn-link"},{icon:p(()=>[f(t(T),{icon:"gearFilled",class:"k-button-icon",size:"16",color:"currentColor","hide-title":""})]),default:p(()=>[i(),Ye]),_:1})])):M("",!0)])}}});const He=k(Ge,[["__scopeId","data-v-c348c723"]]),We=""+new URL("kuma-loader-v1-2aaed7d4.gif",import.meta.url).href,Qe=a=>(x("data-v-06e19708"),a=a(),O(),a),je={class:"full-screen"},Ze={class:"loading-container"},Je=Qe(()=>o("img",{src:We},null,-1)),Xe={class:"progress"},et=b({__name:"AppLoadingBar",setup(a){let e;const s=U(10);return z(function(){e=window.setInterval(()=>{s.value>=100&&(window.clearInterval(e),s.value=100),s.value=Math.min(s.value+Math.ceil(Math.random()*30),100)},150)}),W(function(){window.clearInterval(e)}),(l,c)=>(r(),_("div",je,[o("div",Ze,[Je,i(),o("div",Xe,[o("div",{style:te({width:`${s.value}%`}),class:"progress-bar",role:"progressbar","data-testid":"app-progress-bar"},null,4)])])]))}});const tt=k(et,[["__scopeId","data-v-06e19708"]]),st={key:0,class:"onboarding-check"},ot={class:"alert-content"},nt=b({__name:"AppOnboardingNotification",setup(a){const e=U(!1);function s(){e.value=!0}return(l,c)=>e.value===!1?(r(),_("div",st,[f(t(G),{appearance:"success",class:"dismissible","dismiss-type":"icon",onClosed:s},{alertMessage:p(()=>[o("div",ot,[o("div",null,[o("strong",null,"Welcome to "+h(t(ge))+"!",1),i(` We've detected that you don't have any data plane proxies running yet. We've created an onboarding process to help you!
          `)]),i(),o("div",null,[f(t(I),{appearance:"primary",size:"small",class:"action-button",to:{name:"onboarding-welcome"}},{default:p(()=>[i(`
              Get started
            `)]),_:1})])])]),_:1})])):M("",!0)}});const at=k(nt,[["__scopeId","data-v-c21dc5a7"]]);async function it(a,e,s=()=>!1){do{if(await a(),await s())break;const l=typeof e=="number"?e:e();await new Promise(c=>setTimeout(c,Math.max(0,l)))}while(!await s())}const ct=a=>(x("data-v-76b8351f"),a=a(),O(),a),rt={class:"mesh-selector-container"},lt={for:"mesh-selector"},ut=ct(()=>o("span",{class:"visually-hidden"},`
        Filter by mesh:
      `,-1)),_t=["value","selected"],dt=b({__name:"AppMeshSelector",props:{meshes:{type:Array,required:!0}},setup(a){const e=a,s=K(),l=Y(),c=$(),n=A(()=>c.state.selectedMesh===null?e.meshes[0].name:c.state.selectedMesh);function d(u){const m=u.target.value;c.dispatch("updateSelectedMesh",m);const y="mesh"in s.params?s.name:"mesh-detail-view";l.push({name:y,params:{mesh:m}})}return(u,v)=>(r(),_("div",rt,[o("label",lt,[ut,i(),o("select",{id:"mesh-selector",class:"mesh-selector",name:"mesh-selector","data-testid":"mesh-selector",onChange:d},[(r(!0),_(N,null,q(e.meshes,m=>(r(),_("option",{key:m.name,value:m.name,selected:m.name===t(n)},h(m.name),9,_t))),128))],32)])]))}});const pt=k(dt,[["__scopeId","data-v-76b8351f"]]),mt=["data-testid"],ft={key:1,class:"nav-category"},ht=b({__name:"AppNavItem",props:{name:{type:String,required:!0},routeName:{type:String,required:!1,default:""},usesMeshParam:{type:Boolean,required:!1,default:!1},categoryTier:{type:String,required:!1,default:null},insightsFieldAccessor:{type:String,required:!1,default:""},shouldOffsetFromFollowingItems:{type:Boolean,required:!1,default:!1}},setup(a){const e=a,s=K(),l=Y(),c=$(),n=A(()=>{if(e.insightsFieldAccessor){const m=J(c.state.sidebar.insights,e.insightsFieldAccessor,0);return m>99?"99+":String(m)}else return""}),d=A(()=>{if(e.routeName==="")return null;const m={name:e.routeName};return e.usesMeshParam&&(m.params={mesh:c.state.selectedMesh}),m}),u=A(()=>{if(d.value===null)return!1;if(e.routeName===s.name||s.path.split("/")[2]===d.value.name)return!0;if(s.meta.parent)try{if(l.resolve({name:s.meta.parent}).name===e.routeName)return!0}catch(y){if(y instanceof Error&&y.message.includes("No match for"))console.warn(y);else throw y}return e.routeName&&s.matched.some(y=>e.routeName===y.name||e.routeName===y.redirect)});function v(){X.logger.info(ve.SIDEBAR_ITEM_CLICKED,{data:d.value})}return(m,y)=>{const R=V("router-link");return r(),_("div",{class:D(["nav-item",{"nav-item--is-category":t(d)===null,"nav-item--has-bottom-offset":e.shouldOffsetFromFollowingItems,[`nav-item--is-${e.categoryTier}-category`]:e.categoryTier!==null}]),"data-testid":e.routeName},[t(d)!==null?(r(),g(R,{key:0,class:D(["nav-link",{"nav-link--is-active":t(u)}]),to:t(d),onClick:v},{default:p(()=>[i(h(a.name)+" ",1),t(n)?(r(),_("span",{key:0,class:D(["amount",{"amount--empty":t(n)==="0"}])},h(t(n)),3)):M("",!0)]),_:1},8,["class","to"])):(r(),_("div",ft,h(a.name),1))],10,mt)}}});const gt=k(ht,[["__scopeId","data-v-938e565b"]]),vt={class:"app-sidebar-wrapper"},yt={class:"app-sidebar"},bt=b({__name:"AppSidebar",setup(a){const s=$(),l=he(),c=A(()=>l(s.getters["config/getMulticlusterStatus"],s.state.meshes.items.length>0)),n=A(()=>s.state.meshes.items);E(()=>s.state.selectedMesh,()=>{s.dispatch("sidebar/getMeshInsights")});let d=!1;z(function(){window.addEventListener("blur",u),window.addEventListener("focus",v)}),W(function(){window.removeEventListener("blur",u),window.removeEventListener("focus",v)}),v();function u(){d=!0}function v(){d=!1,it(m,10*1e3,()=>d)}function m(){return s.dispatch("sidebar/getInsights")}return(y,R)=>(r(),_("div",vt,[o("aside",yt,[(r(!0),_(N,null,q(t(c),(C,j)=>(r(),_(N,{key:j},[C.isMeshSelector?(r(),_(N,{key:0},[t(n).length>0?(r(),g(pt,{key:0,meshes:t(n)},null,8,["meshes"])):M("",!0)],64)):(r(),g(gt,se(oe({key:1},C)),null,16))],64))),128))])]))}});const At=k(bt,[["__scopeId","data-v-ddb44585"]]),Mt={class:"py-4"},kt=o("p",{class:"mb-4"},`
      A traffic log policy lets you collect access logs for every data plane proxy in your service mesh.
    `,-1),$t={class:"list-disc pl-4"},St=["href"],wt=b({__name:"LoggingNotification",setup(a){const e=L();return(s,l)=>(r(),_("div",Mt,[kt,i(),o("ul",$t,[o("li",null,[o("a",{href:`${t(e)("KUMA_DOCS_URL")}/policies/traffic-log/?${t(e)("KUMA_UTM_QUERY_PARAMS")}`,target:"_blank"},`
          Traffic Log policy documentation
        `,8,St)])])]))}}),Nt={class:"py-4"},Ut=o("p",{class:"mb-4"},`
      A traffic metrics policy lets you collect key data for observability of your service mesh.
    `,-1),It={class:"list-disc pl-4"},Lt=["href"],Rt=b({__name:"MetricsNotification",setup(a){const e=L();return(s,l)=>(r(),_("div",Nt,[Ut,i(),o("ul",It,[o("li",null,[o("a",{href:`${t(e)("KUMA_DOCS_URL")}/policies/traffic-metrics/?${t(e)("KUMA_UTM_QUERY_PARAMS")}`,target:"_blank"},`
          Traffic Metrics policy documentation
        `,8,Lt)])])]))}}),Tt={class:"py-4"},Ct=o("p",{class:"mb-4"},`
      Mutual TLS (mTLS) for communication between all the components
      of your service mesh (services, control plane, data plane proxies), proxy authentication,
      and access control rules in Traffic Permissions policies all contribute to securing your mesh.
    `,-1),Et={class:"list-disc pl-4"},Pt=["href"],xt=["href"],Ot=["href"],Kt=b({__name:"MtlsNotification",setup(a){const e=L();return(s,l)=>(r(),_("div",Tt,[Ct,i(),o("ul",Et,[o("li",null,[o("a",{href:`${t(e)("KUMA_DOCS_URL")}/security/certificates/?${t(e)("KUMA_UTM_QUERY_PARAMS")}`,target:"_blank"},`
          Secure access across services
        `,8,Pt)]),i(),o("li",null,[o("a",{href:`${t(e)("KUMA_DOCS_URL")}/policies/mutual-tls/?${t(e)("KUMA_UTM_QUERY_PARAMS")}`,target:"_blank"},`
          Mutual TLS
        `,8,xt)]),i(),o("li",null,[o("a",{href:`${t(e)("KUMA_DOCS_URL")}/policies/traffic-permissions/?${t(e)("KUMA_UTM_QUERY_PARAMS")}`,target:"_blank"},`
          Traffic Permissions policy documentation
        `,8,Ot)])])]))}}),Dt={class:"py-4"},Bt=o("p",{class:"mb-4"},`
      A traffic trace policy lets you enable tracing logs and a third-party tracing solution to send them to.
    `,-1),Ft={class:"list-disc pl-4"},Vt=["href"],zt=b({__name:"TracingNotification",setup(a){const e=L();return(s,l)=>(r(),_("div",Dt,[Bt,i(),o("ul",Ft,[o("li",null,[o("a",{href:`${t(e)("KUMA_DOCS_URL")}/policies/traffic-trace/?${t(e)("KUMA_UTM_QUERY_PARAMS")}`,target:"_blank"},`
          Traffic Trace policy documentation
        `,8,Vt)])])]))}}),qt={class:"flex items-center"},Yt=b({__name:"SingleMeshNotifications",setup(a){const e=$(),s={LoggingNotification:wt,MetricsNotification:Rt,MtlsNotification:Kt,TracingNotification:zt};return(l,c)=>(r(),g(be,{"multiple-open":""},{default:p(()=>[(r(!0),_(N,null,q(t(e).getters["notifications/singleMeshNotificationItems"],n=>(r(),g(ye,{key:n.name},{"accordion-header":p(()=>[o("div",qt,[n.isCompleted?(r(),g(t(T),{key:0,color:"var(--green-500)",icon:"check",size:"20",class:"mr-4"})):(r(),g(t(T),{key:1,icon:"warning",color:"var(--black-75)","secondary-color":"var(--yellow-300)",size:"20",class:"mr-4"})),i(),o("strong",null,h(n.name),1)])]),"accordion-content":p(()=>[n.component?(r(),g(Q(s[n.component]),{key:0})):(r(),g(t(de),{key:1},{body:p(()=>[i(h(n.content),1)]),_:2},1024))]),_:2},1024))),128))]),_:1}))}}),Gt=a=>(x("data-v-baf26e82"),a=a(),O(),a),Ht={class:"mr-4"},Wt=Gt(()=>o("span",{class:"mr-2"},[o("strong",null,"Pro tip:"),i(`

            You might want to adjust your mesh configuration
          `)],-1)),Qt={key:0},jt={class:"text-xl tracking-wide"},Zt={key:1},Jt={class:"text-xl tracking-wide"},Xt=b({__name:"NotificationManager",setup(a){const e=$(),s=U(!0),l=A(()=>e.state.selectedMesh?e.getters["notifications/meshNotificationItemMapWithAction"][e.state.selectedMesh]:!1);z(function(){const u=B.get("hideCheckMeshAlert");s.value=u!=="yes"});function c(){s.value=!1,B.set("hideCheckMeshAlert","yes")}function n(){e.dispatch("notifications/openModal")}function d(){e.dispatch("notifications/closeModal")}return(u,v)=>(r(),_("div",null,[s.value?(r(),g(t(G),{key:0,class:"mb-4",appearance:"info","dismiss-type":"icon","data-testid":"notification-info",onClosed:c},{alertMessage:p(()=>[o("div",Ht,[Wt,i(),f(t(I),{appearance:"outline","data-testid":"open-modal-button",onClick:n},{default:p(()=>[i(`
            Check your mesh!
          `)]),_:1})])]),_:1})):M("",!0),i(),f(t(pe),{class:"modal","is-visible":t(e).state.notifications.isOpen,title:"Notifications","text-align":"left","data-testid":"notification-modal"},{"header-content":p(()=>[o("div",null,[o("div",null,[t(l)?(r(),_("span",Qt,[i(`
              Some of these features are not enabled for `),o("span",jt,'"'+h(t(e).state.selectedMesh)+'"',1),i(` mesh. Consider implementing them.
            `)])):(r(),_("span",Zt,[i(`
              Looks like `),o("span",Jt,'"'+h(t(e).state.selectedMesh)+'"',1),i(` isn't missing any features. Well done!
            `)]))])])]),"body-content":p(()=>[f(Yt)]),"footer-content":p(()=>[f(t(I),{appearance:"outline","data-testid":"close-modal-button",onClick:d},{default:p(()=>[i(`
          Close
        `)]),_:1})]),_:1},8,["is-visible"])]))}});const es=k(Xt,[["__scopeId","data-v-baf26e82"]]),ts={key:0},ss={key:1,class:"app-content-container"},os={class:"app-main-content"},ns=b({__name:"App",setup(a){const e=$(),s=K(),l=U(e.state.globalLoading),c=A(()=>s.path),n=A(()=>e.state.config.status!=="OK"),d=A(()=>e.getters.shouldSuggestOnboarding),u=A(()=>e.getters["notifications/amountOfActions"]>0);E(()=>e.state.globalLoading,function(m){l.value=m}),E(()=>s.meta.title,function(m){v(m)}),E(()=>e.state.pageTitle,function(m){v(m)});function v(m){const y="Kuma Manager";document.title=m?`${m} | ${y}`:y}return(m,y)=>{const R=V("router-view");return l.value?(r(),g(tt,{key:0})):(r(),_(N,{key:1},[f(He),i(),t(s).meta.onboardingProcess?(r(),_("div",ts,[f(R)])):(r(),_("div",ss,[f(At),i(),o("main",os,[t(n)?(r(),g(Ue,{key:0})):M("",!0),i(),t(u)?(r(),g(es,{key:1})):M("",!0),i(),t(d)?(r(),g(at,{key:2})):M("",!0),i(),f(Se),i(),(r(),g(R,{key:t(c)},{default:p(({Component:C})=>[f(ne,{mode:"out-in",name:"fade"},{default:p(()=>[(r(),_("div",{key:t(s).name,class:"transition-root"},[(r(),g(Q(C)))]))]),_:2},1024)]),_:1}))])]))],64))}}});const as=k(ns,[["__scopeId","data-v-50928ee3"]]);async function is(a,e,s){const l=S(w.store),c=S(w.api);if(document.title=`${a("KUMA_PRODUCT_NAME")} Manager`,c.setBaseUrl(a("KUMA_API_URL")),{}.VITE_MOCK_API_ENABLED==="true"){const{setupMockWorker:u}=await ee(()=>import("./setupMockWorker-d70c3c41.js"),[],import.meta.url);u(c.baseUrl)}(async()=>{const u=await c.getConfig();s.setup(u)})();const n=ae(as);n.use(l,S(w.storeKey)),await Promise.all([l.dispatch("bootstrap"),l.dispatch("fetchPolicyTypes")]);const d=await Ae(e,a("KUMA_BASE_PATH"));n.use(d),n.mount("#app")}is(S(w.env),S(w.routes),S(w.logger));
