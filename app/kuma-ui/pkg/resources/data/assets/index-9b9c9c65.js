import{u as R,a as Q,P as z,C as N,g as W,T as q}from"./production-060535a4.js";import{d as f,c as v,u as t,o as a,a as u,b as y,w as _,e as m,f as i,g as s,h as p,t as b,r as A,i as O,j as V,n as F,p as P,k as K,F as I,l as j,m as E,q as U,s as H,T as G}from"./runtime-dom.esm-bundler-062436f2.js";import{u as x}from"./store-3df31b4b.js";import{q as J,m as S,T as X,L as w,O as D,_ as Z,a as ee}from"./kongponents.es-79677c68.js";import{u as te,a as M}from"./index-c2dc68c3.js";import{_ as $}from"./_plugin-vue_export-helper-c27b6911.js";import{A as se,a as oe}from"./AccordionItem-a2656bd1.js";import{u as ne,a as ae}from"./index-268bbdd9.js";import"./DoughnutChart-bdf94136.js";import"./datadogLogEvents-302eea7b.js";(function(){const e=document.createElement("link").relList;if(e&&e.supports&&e.supports("modulepreload"))return;for(const n of document.querySelectorAll('link[rel="modulepreload"]'))r(n);new MutationObserver(n=>{for(const o of n)if(o.type==="childList")for(const g of o.addedNodes)g.tagName==="LINK"&&g.rel==="modulepreload"&&r(g)}).observe(document,{childList:!0,subtree:!0});function c(n){const o={};return n.integrity&&(o.integrity=n.integrity),n.referrerpolicy&&(o.referrerPolicy=n.referrerpolicy),n.crossorigin==="use-credentials"?o.credentials="include":n.crossorigin==="anonymous"?o.credentials="omit":o.credentials="same-origin",o}function r(n){if(n.ep)return;n.ep=!0;const o=c(n);fetch(n.href,o)}})();const ie=f({__name:"AppBreadcrumbs",setup(l){const e=R(),c=Q(),r=v(()=>{const n=new Map;for(const o of e.matched){if(o.name==="home"||o.meta.parent==="home")continue;if(o.meta.parent!==void 0){const d=c.resolve({name:o.meta.parent});d.name&&n.set(d.name,{to:d,key:d.name,title:d.meta.title,text:d.meta.title})}if((o.name===e.name||o.redirect===e.name)&&o.meta.breadcrumbExclude!==!0&&e.name){let d=e.meta.title;e.meta.breadcrumbTitleParam&&e.params[e.meta.breadcrumbTitleParam]&&(d=e.params[e.meta.breadcrumbTitleParam]),n.set(e.name,{to:e,key:e.name,title:d,text:d})}}return Array.from(n.values())});return(n,o)=>t(r).length>0?(a(),u(t(J),{key:0,items:t(r)},null,8,["items"])):y("",!0)}}),ce=s("p",null,"Unable to reach the API",-1),re={key:0},le=f({__name:"AppErrorMessage",setup(l){const e=te();return(c,r)=>(a(),u(t(X),{class:"global-api-status empty-state--wide-content empty-state--compact","cta-is-hidden":""},{title:_(()=>[m(t(S),{class:"mb-3",icon:"warning",color:"var(--black-75)","secondary-color":"var(--yellow-300)",size:"64"}),i(),ce]),message:_(()=>[s("p",null,[i(`
        Please double check to make sure it is up and running `),t(e).baseUrl?(a(),p("span",re,[i(", and it is reachable at "),s("code",null,b(t(e).baseUrl),1)])):y("",!0)])]),_:1}))}}),_e=""+new URL("kuma-loader-v1-2aaed7d4.gif",import.meta.url).href,de=l=>(P("data-v-06e19708"),l=l(),K(),l),ue={class:"full-screen"},pe={class:"loading-container"},me=de(()=>s("img",{src:_e},null,-1)),fe={class:"progress"},he=f({__name:"AppLoadingBar",setup(l){let e;const c=A(10);return O(function(){e=window.setInterval(()=>{c.value>=100&&(window.clearInterval(e),c.value=100),c.value=Math.min(c.value+Math.ceil(Math.random()*30),100)},150)}),V(function(){window.clearInterval(e)}),(r,n)=>(a(),p("div",ue,[s("div",pe,[me,i(),s("div",fe,[s("div",{style:F({width:`${c.value}%`}),class:"progress-bar",role:"progressbar","data-testid":"app-progress-bar"},null,4)])])]))}});const ge=$(he,[["__scopeId","data-v-06e19708"]]),ye={key:0,class:"onboarding-check"},ve={class:"alert-content"},be=f({__name:"AppOnboardingNotification",setup(l){const e=A(!1);function c(){e.value=!0}return(r,n)=>e.value===!1?(a(),p("div",ye,[m(t(D),{appearance:"success",class:"dismissible","dismiss-type":"icon",onClosed:c},{alertMessage:_(()=>[s("div",ve,[s("div",null,[s("strong",null,"Welcome to "+b(t(z))+"!",1),i(` We've detected that you don't have any data plane proxies running yet. We've created an onboarding process to help you!
          `)]),i(),s("div",null,[m(t(w),{appearance:"primary",size:"small",class:"action-button",to:{name:"onboarding-welcome"}},{default:_(()=>[i(`
              Get started
            `)]),_:1})])])]),_:1})])):y("",!0)}});const Ae=$(be,[["__scopeId","data-v-c21dc5a7"]]),Me={class:"py-4"},$e=s("p",{class:"mb-4"},`
      A traffic log policy lets you collect access logs for every data plane proxy in your service mesh.
    `,-1),ke={class:"list-disc pl-4"},Ue=["href"],Se=f({__name:"LoggingNotification",setup(l){const e=M();return(c,r)=>(a(),p("div",Me,[$e,i(),s("ul",ke,[s("li",null,[s("a",{href:`${t(e)("KUMA_DOCS_URL")}/policies/traffic-log/?${t(e)("KUMA_UTM_QUERY_PARAMS")}`,target:"_blank"},`
          Traffic Log policy documentation
        `,8,Ue)])])]))}}),we={class:"py-4"},xe=s("p",{class:"mb-4"},`
      A traffic metrics policy lets you collect key data for observability of your service mesh.
    `,-1),Te={class:"list-disc pl-4"},Le=["href"],Ce=f({__name:"MetricsNotification",setup(l){const e=M();return(c,r)=>(a(),p("div",we,[xe,i(),s("ul",Te,[s("li",null,[s("a",{href:`${t(e)("KUMA_DOCS_URL")}/policies/traffic-metrics/?${t(e)("KUMA_UTM_QUERY_PARAMS")}`,target:"_blank"},`
          Traffic Metrics policy documentation
        `,8,Le)])])]))}}),Ne={class:"py-4"},Re=s("p",{class:"mb-4"},`
      Mutual TLS (mTLS) for communication between all the components
      of your service mesh (services, control plane, data plane proxies), proxy authentication,
      and access control rules in Traffic Permissions policies all contribute to securing your mesh.
    `,-1),Oe={class:"list-disc pl-4"},Pe=["href"],Ke=["href"],Ie=["href"],Ee=f({__name:"MtlsNotification",setup(l){const e=M();return(c,r)=>(a(),p("div",Ne,[Re,i(),s("ul",Oe,[s("li",null,[s("a",{href:`${t(e)("KUMA_DOCS_URL")}/security/certificates/?${t(e)("KUMA_UTM_QUERY_PARAMS")}`,target:"_blank"},`
          Secure access across services
        `,8,Pe)]),i(),s("li",null,[s("a",{href:`${t(e)("KUMA_DOCS_URL")}/policies/mutual-tls/?${t(e)("KUMA_UTM_QUERY_PARAMS")}`,target:"_blank"},`
          Mutual TLS
        `,8,Ke)]),i(),s("li",null,[s("a",{href:`${t(e)("KUMA_DOCS_URL")}/policies/traffic-permissions/?${t(e)("KUMA_UTM_QUERY_PARAMS")}`,target:"_blank"},`
          Traffic Permissions policy documentation
        `,8,Ie)])])]))}}),De={class:"py-4"},Be=s("p",{class:"mb-4"},`
      A traffic trace policy lets you enable tracing logs and a third-party tracing solution to send them to.
    `,-1),Ye={class:"list-disc pl-4"},Qe=["href"],ze=f({__name:"TracingNotification",setup(l){const e=M();return(c,r)=>(a(),p("div",De,[Be,i(),s("ul",Ye,[s("li",null,[s("a",{href:`${t(e)("KUMA_DOCS_URL")}/policies/traffic-trace/?${t(e)("KUMA_UTM_QUERY_PARAMS")}`,target:"_blank"},`
          Traffic Trace policy documentation
        `,8,Qe)])])]))}}),We={class:"flex items-center"},qe=f({__name:"SingleMeshNotifications",setup(l){const e=x(),c={LoggingNotification:Se,MetricsNotification:Ce,MtlsNotification:Ee,TracingNotification:ze};return(r,n)=>(a(),u(oe,{"multiple-open":""},{default:_(()=>[(a(!0),p(I,null,j(t(e).getters["notifications/singleMeshNotificationItems"],o=>(a(),u(se,{key:o.name},{"accordion-header":_(()=>[s("div",We,[o.isCompleted?(a(),u(t(S),{key:0,color:"var(--green-500)",icon:"check",size:"20",class:"mr-4"})):(a(),u(t(S),{key:1,icon:"warning",color:"var(--black-75)","secondary-color":"var(--yellow-300)",size:"20",class:"mr-4"})),i(),s("strong",null,b(o.name),1)])]),"accordion-content":_(()=>[o.component?(a(),u(E(c[o.component]),{key:0})):(a(),u(t(Z),{key:1},{body:_(()=>[i(b(o.content),1)]),_:2},1024))]),_:2},1024))),128))]),_:1}))}}),Ve=l=>(P("data-v-baf26e82"),l=l(),K(),l),Fe={class:"mr-4"},je=Ve(()=>s("span",{class:"mr-2"},[s("strong",null,"Pro tip:"),i(`

            You might want to adjust your mesh configuration
          `)],-1)),He={key:0},Ge={class:"text-xl tracking-wide"},Je={key:1},Xe={class:"text-xl tracking-wide"},Ze=f({__name:"NotificationManager",setup(l){const e=x(),c=A(!0),r=v(()=>e.state.selectedMesh?e.getters["notifications/meshNotificationItemMapWithAction"][e.state.selectedMesh]:!1);O(function(){const d=N.get("hideCheckMeshAlert");c.value=d!=="yes"});function n(){c.value=!1,N.set("hideCheckMeshAlert","yes")}function o(){e.dispatch("notifications/openModal")}function g(){e.dispatch("notifications/closeModal")}return(d,T)=>(a(),p("div",null,[c.value?(a(),u(t(D),{key:0,class:"mb-4",appearance:"info","dismiss-type":"icon","data-testid":"notification-info",onClosed:n},{alertMessage:_(()=>[s("div",Fe,[je,i(),m(t(w),{appearance:"outline","data-testid":"open-modal-button",onClick:o},{default:_(()=>[i(`
            Check your mesh!
          `)]),_:1})])]),_:1})):y("",!0),i(),m(t(ee),{class:"modal","is-visible":t(e).state.notifications.isOpen,title:"Notifications","text-align":"left","data-testid":"notification-modal"},{"header-content":_(()=>[s("div",null,[s("div",null,[t(r)?(a(),p("span",He,[i(`
              Some of these features are not enabled for `),s("span",Ge,'"'+b(t(e).state.selectedMesh)+'"',1),i(` mesh. Consider implementing them.
            `)])):(a(),p("span",Je,[i(`
              Looks like `),s("span",Xe,'"'+b(t(e).state.selectedMesh)+'"',1),i(` isn't missing any features. Well done!
            `)]))])])]),"body-content":_(()=>[m(qe)]),"footer-content":_(()=>[m(t(w),{appearance:"outline","data-testid":"close-modal-button",onClick:g},{default:_(()=>[i(`
          Close
        `)]),_:1})]),_:1},8,["is-visible"])]))}});const et=$(Ze,[["__scopeId","data-v-baf26e82"]]),tt={key:0},st={key:1,class:"app-content-container"},ot={class:"app-main-content"},nt=f({__name:"App",setup(l){const[e,c]=[ne(),ae()],r=x(),n=R(),o=A(r.state.globalLoading),g=v(()=>n.path),d=v(()=>r.state.config.status!=="OK"),T=v(()=>r.getters.shouldSuggestOnboarding),B=v(()=>r.getters["notifications/amountOfActions"]>0);U(()=>r.state.globalLoading,function(h){o.value=h}),U(()=>n.meta.title,function(h){L(h)}),U(()=>r.state.pageTitle,function(h){L(h)});function L(h){const k="Kuma Manager";document.title=h?`${h} | ${k}`:k}return(h,k)=>{const C=H("router-view");return o.value?(a(),u(ge,{key:0})):(a(),p(I,{key:1},[m(t(c)),i(),t(n).meta.onboardingProcess?(a(),p("div",tt,[m(C)])):(a(),p("div",st,[m(t(e)),i(),s("main",ot,[t(d)?(a(),u(le,{key:0})):y("",!0),i(),t(B)?(a(),u(et,{key:1})):y("",!0),i(),t(T)?(a(),u(Ae,{key:2})):y("",!0),i(),m(ie),i(),(a(),u(C,{key:t(g)},{default:_(({Component:Y})=>[m(G,{mode:"out-in",name:"fade"},{default:_(()=>[(a(),p("div",{key:t(n).name,class:"transition-root"},[(a(),u(E(Y)))]))]),_:2},1024)]),_:1}))])]))],64))}}});const at=$(nt,[["__scopeId","data-v-fe495b2f"]]);(async()=>await W(q.app)(at))();
