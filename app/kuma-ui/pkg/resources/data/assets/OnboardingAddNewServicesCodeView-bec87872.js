import{L as y}from"./LoadingBox-8fd90a16.js";import{O as A,a as V,b as N}from"./OnboardingPage-21ee557e.js";import{_ as m}from"./CodeBlock.vue_vue_type_style_index_0_lang-226d1ddf.js";import{d as T,u as C,m as _,S as O,a as r,o,b as R,w as a,e as t,f as e,c,F as P,p as n,v as D,A as I,B as L,_ as S}from"./index-784d2bbf.js";const b=d=>(I("data-v-ad84ed8e"),d=d(),L(),d),B=b(()=>n("p",{class:"mb-4 text-center"},`
            The demo application includes two services: a Redis backend to store a counter value, and a frontend web UI to show and increment the counter.
          `,-1)),E=b(()=>n("p",null,"To run execute the following command:",-1)),G={key:1},q={class:"status-box mt-4"},F={key:0,class:"status--is-connected","data-testid":"dpps-connected"},H={key:1,class:"status--is-disconnected","data-testid":"dpps-disconnected"},K={key:0,class:"status-loading-box mt-4"},M=1e3,h="https://github.com/kumahq/kuma-counter-demo/",U="https://github.com/kumahq/kuma-counter-demo/blob/master/README.md",$="kubectl apply -f https://bit.ly/3Kh2Try",j=T({__name:"OnboardingAddNewServicesCodeView",setup(d){const g=C(),s=_(!1),l=_(null);u(),O(function(){p()});async function u(){try{const{total:i}=await g.getAllDataplanes();s.value=i>0}catch(i){console.error(i)}finally{s.value||(p(),l.value=window.setTimeout(()=>u(),M))}}function p(){l.value!==null&&window.clearTimeout(l.value)}return(i,z)=>{const v=r("RouteTitle"),w=r("AppView"),f=r("RouteView");return o(),R(f,{name:"onboarding-add-new-services"},{default:a(({can:k,t:x})=>[t(v,{title:x("onboarding.routes.add-services-code.title"),render:!1},null,8,["title"]),e(),t(w,null,{default:a(()=>[t(A,null,{header:a(()=>[t(V,null,{title:a(()=>[e(`
              Add services
            `)]),_:1})]),content:a(()=>[B,e(),k("use kubernetes")?(o(),c(P,{key:0},[E,e(),t(m,{id:"code-block-kubernetes-command",language:"bash",code:$})],64)):(o(),c("div",G,[n("p",{class:"mb-4 text-center"},[e(`
              Clone `),n("a",{href:h,target:"_blank"},"the GitHub repository"),e(` for the demo application:
            `)]),e(),t(m,{id:"code-block-clone-command",language:"bash",code:`git clone ${h}`},null,8,["code"]),e(),n("p",{class:"mt-4 text-center"},[e(`
              And follow the instructions in `),n("a",{href:U,target:"_blank"},"the README"),e(`.
            `)])])),e(),n("div",null,[n("p",q,[e(`
              DPPs status:

              `),s.value?(o(),c("span",F,"Connected")):(o(),c("span",H,"Disconnected"))]),e(),s.value?D("",!0):(o(),c("div",K,[t(y)]))])]),navigation:a(()=>[t(N,{"next-step":"onboarding-dataplanes-view","previous-step":"onboarding-add-new-services-view","should-allow-next":s.value},null,8,["should-allow-next"])]),_:2},1024)]),_:2},1024)]),_:1})}}});const Y=S(j,[["__scopeId","data-v-ad84ed8e"]]);export{Y as default};
