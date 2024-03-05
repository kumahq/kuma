import{L as y}from"./LoadingBox-Z3sh0QjM.js";import{O as A,a as C,b as R}from"./OnboardingPage-FM9NRJTr.js";import{C as m}from"./CodeBlock-6c7dCnil.js";import{d as T,Q as V,C as _,R as N,a as r,o,b as O,w as a,e as t,f as e,F as P,c,m as n,p as D,x as I,y as L,_ as B}from"./index-pAyRVwwQ.js";const b=i=>(I("data-v-34beecb4"),i=i(),L(),i),E=b(()=>n("p",{class:"mb-4 text-center"},`
            The demo application includes two services: a Redis backend to store a counter value, and a frontend web UI to show and increment the counter.
          `,-1)),S=b(()=>n("p",null,"To run execute the following command:",-1)),G={key:1},q={class:"status-box mt-4"},F={key:0,class:"status--is-connected","data-testid":"dpps-connected"},H={key:1,class:"status--is-disconnected","data-testid":"dpps-disconnected"},K={key:0,class:"status-loading-box mt-4"},M=1e3,h="https://github.com/kumahq/kuma-counter-demo/",U="https://github.com/kumahq/kuma-counter-demo/blob/master/README.md",Q="kubectl apply -f https://bit.ly/3Kh2Try",$=T({__name:"OnboardingAddNewServicesCodeView",setup(i){const g=V(),s=_(!1),l=_(null);u(),N(function(){p()});async function u(){try{const{total:d}=await g.getAllDataplanes();s.value=d>0}catch(d){console.error(d)}finally{s.value||(p(),l.value=window.setTimeout(()=>u(),M))}}function p(){l.value!==null&&window.clearTimeout(l.value)}return(d,j)=>{const f=r("RouteTitle"),v=r("AppView"),w=r("RouteView");return o(),O(w,{name:"onboarding-add-new-services"},{default:a(({can:k,t:x})=>[t(f,{title:x("onboarding.routes.add-services-code.title"),render:!1},null,8,["title"]),e(),t(v,null,{default:a(()=>[t(A,null,{header:a(()=>[t(C,null,{title:a(()=>[e(`
              Add services
            `)]),_:1})]),content:a(()=>[E,e(),k("use kubernetes")?(o(),c(P,{key:0},[S,e(),t(m,{language:"bash",code:Q})],64)):(o(),c("div",G,[n("p",{class:"mb-4 text-center"},[e(`
              Clone `),n("a",{href:h,target:"_blank"},"the GitHub repository"),e(` for the demo application:
            `)]),e(),t(m,{language:"bash",code:`git clone ${h}`},null,8,["code"]),e(),n("p",{class:"mt-4 text-center"},[e(`
              And follow the instructions in `),n("a",{href:U,target:"_blank"},"the README"),e(`.
            `)])])),e(),n("div",null,[n("p",q,[e(`
              DPPs status:

              `),s.value?(o(),c("span",F,"Connected")):(o(),c("span",H,"Disconnected"))]),e(),s.value?D("",!0):(o(),c("div",K,[t(y)]))])]),navigation:a(()=>[t(R,{"next-step":"onboarding-dataplanes-view","previous-step":"onboarding-add-new-services-view","should-allow-next":s.value},null,8,["should-allow-next"])]),_:2},1024)]),_:2},1024)]),_:1})}}}),Y=B($,[["__scopeId","data-v-34beecb4"]]);export{Y as default};
