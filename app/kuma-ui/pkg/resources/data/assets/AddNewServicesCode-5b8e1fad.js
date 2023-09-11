import{L as v}from"./LoadingBox-78e671cd.js";import{O as k,a as w,b as x}from"./OnboardingPage-e9d9a39b.js";import{d as y,L as A,R as N,t as m,S as C,o as s,g as T,w as a,h as t,i as I,C as L,l as e,j as c,F as P,m as n,k as D,A as O,_ as R,z as S,B,q as E}from"./index-c11fbf03.js";import{_ as p}from"./CodeBlock.vue_vue_type_style_index_0_lang-d61f09bd.js";const h=d=>(S("data-v-53d3620d"),d=d(),B(),d),V=h(()=>n("p",{class:"mb-4 text-center"},`
            The demo application includes two services: a Redis backend to store a counter value, and a frontend web UI to show and increment the counter.
          `,-1)),$=h(()=>n("p",null,"To run execute the following command:",-1)),q={key:1},G={class:"status-box mt-4"},F={key:0,class:"status--is-connected","data-testid":"dpps-connected"},H={key:1,class:"status--is-disconnected","data-testid":"dpps-disconnected"},K={key:0,class:"status-loading-box mt-4"},M=1e3,_="https://github.com/kumahq/kuma-counter-demo/",U="https://github.com/kumahq/kuma-counter-demo/blob/master/README.md",j="kubectl apply -f https://bit.ly/3Kh2Try",z=y({__name:"AddNewServicesCode",setup(d){const{t:b}=A(),f=N(),o=m(!1),l=m(null);u(),C(function(){r()});async function u(){try{const{total:i}=await f.getAllDataplanes();o.value=i>0}catch(i){console.error(i)}finally{o.value||(r(),l.value=window.setTimeout(()=>u(),M))}}function r(){l.value!==null&&window.clearTimeout(l.value)}return(i,J)=>(s(),T(R,null,{default:a(({can:g})=>[t(L,{title:I(b)("onboarding.routes.add-services-code.title")},null,8,["title"]),e(),t(O,null,{default:a(()=>[t(k,null,{header:a(()=>[t(w,null,{title:a(()=>[e(`
              Add services
            `)]),_:1})]),content:a(()=>[V,e(),g("use kubernetes")?(s(),c(P,{key:0},[$,e(),t(p,{id:"code-block-kubernetes-command",language:"bash",code:j})],64)):(s(),c("div",q,[n("p",{class:"mb-4 text-center"},[e(`
              Clone `),n("a",{href:_,target:"_blank"},"the GitHub repository"),e(` for the demo application:
            `)]),e(),t(p,{id:"code-block-clone-command",language:"bash",code:`git clone ${_}`},null,8,["code"]),e(),n("p",{class:"mt-4 text-center"},[e(`
              And follow the instructions in `),n("a",{href:U,target:"_blank"},"the README"),e(`.
            `)])])),e(),n("div",null,[n("p",G,[e(`
              DPPs status:

              `),o.value?(s(),c("span",F,"Connected")):(s(),c("span",H,"Disconnected"))]),e(),o.value?D("",!0):(s(),c("div",K,[t(v)]))])]),navigation:a(()=>[t(x,{"next-step":"onboarding-dataplanes-overview","previous-step":"onboarding-add-services","should-allow-next":o.value},null,8,["should-allow-next"])]),_:2},1024)]),_:2},1024)]),_:1}))}});const Z=E(z,[["__scopeId","data-v-53d3620d"]]);export{Z as default};
