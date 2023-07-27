import{L as k}from"./LoadingBox-daaec13d.js";import{O as w,a as x,b as y}from"./OnboardingPage-26d6ff1f.js";import{g as A,n as N,e as C,A as T,_ as I,f as P}from"./RouteView.vue_vue_type_script_setup_true_lang-66b620d6.js";import{_ as D}from"./RouteTitle.vue_vue_type_script_setup_true_lang-6fada8eb.js";import{_ as m}from"./CodeBlock.vue_vue_type_style_index_0_lang-b17fa92f.js";import{d as E,j as p,c as L,x as O,o,a as S,w as s,h as t,b as R,g as e,e as c,F as V,q as n,f as B,p as $,m as q}from"./index-4513d162.js";const h=d=>($("data-v-159bee77"),d=d(),q(),d),G=h(()=>n("p",{class:"mb-4 text-center"},`
            The demo application includes two services: a Redis backend to store a counter value, and a frontend web UI to show and increment the counter.
          `,-1)),K=h(()=>n("p",null,"To run execute the following command:",-1)),F={key:1},H={class:"status-box mt-4"},M={key:0,class:"status--is-connected","data-testid":"dpps-connected"},U={key:1,class:"status--is-disconnected","data-testid":"dpps-disconnected"},j={key:0,class:"status-loading-box mt-4"},z=1e3,_="https://github.com/kumahq/kuma-counter-demo/",J="https://github.com/kumahq/kuma-counter-demo/blob/master/README.md",Q="kubectl apply -f https://bit.ly/3Kh2Try",W=E({__name:"AddNewServicesCode",setup(d){const{t:b}=A(),f=N(),g=C(),a=p(!1),l=p(null),v=L(()=>g.getters["config/getEnvironment"]==="kubernetes");r(),O(function(){u()});async function r(){try{const{total:i}=await f.getAllDataplanes();a.value=i>0}catch(i){console.error(i)}finally{a.value||(u(),l.value=window.setTimeout(()=>r(),z))}}function u(){l.value!==null&&window.clearTimeout(l.value)}return(i,X)=>(o(),S(I,null,{default:s(()=>[t(D,{title:R(b)("onboarding.routes.add-services-code.title")},null,8,["title"]),e(),t(T,null,{default:s(()=>[t(w,null,{header:s(()=>[t(x,null,{title:s(()=>[e(`
              Add services
            `)]),_:1})]),content:s(()=>[G,e(),v.value?(o(),c(V,{key:0},[K,e(),t(m,{id:"code-block-kubernetes-command",language:"bash",code:Q})],64)):(o(),c("div",F,[n("p",{class:"mb-4 text-center"},[e(`
              Clone `),n("a",{href:_,target:"_blank"},"the GitHub repository"),e(` for the demo application:
            `)]),e(),t(m,{id:"code-block-clone-command",language:"bash",code:`git clone ${_}`},null,8,["code"]),e(),n("p",{class:"mt-4 text-center"},[e(`
              And follow the instructions in `),n("a",{href:J,target:"_blank"},"the README"),e(`.
            `)])])),e(),n("div",null,[n("p",H,[e(`
              DPPs status:

              `),a.value?(o(),c("span",M,"Connected")):(o(),c("span",U,"Disconnected"))]),e(),a.value?B("",!0):(o(),c("div",j,[t(k)]))])]),navigation:s(()=>[t(y,{"next-step":"onboarding-dataplanes-overview","previous-step":"onboarding-add-services","should-allow-next":a.value},null,8,["should-allow-next"])]),_:1})]),_:1})]),_:1}))}});const se=P(W,[["__scopeId","data-v-159bee77"]]);export{se as default};
