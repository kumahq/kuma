import{m as _}from"./vuex.esm-bundler-df5bd11e.js";import{j as k}from"./index-a8834e9c.js";import{D as b}from"./kongponents.es-3df60cd6.js";import{k as P}from"./kumaApi-db784568.js";import{P as f}from"./constants-31fdaf55.js";import{k as D}from"./kumaDpServerUrl-b6bb30c6.js";import{_ as v}from"./CodeBlock.vue_vue_type_style_index_0_lang-98064716.js";import{L as x}from"./LoadingBox-695e756c.js";import{O as C}from"./OnboardingNavigation-0bca1fcc.js";import{O as y,a as O}from"./OnboardingPage-0de7feb6.js";import{_ as A}from"./_plugin-vue_export-helper-c27b6911.js";import{l as t,a as N,w as d,o as n,e as a,f as o,h as r,F as T,g as s,t as w,b as L}from"./runtime-dom.esm-bundler-91b41870.js";import"./_commonjsHelpers-87174ba5.js";import"./ClientStorage-efe299d9.js";const B=1e3,R={type:"Dataplane",mesh:"default",name:"example",networking:{address:"localhost",inbound:[{port:7777,servicePort:7777,serviceAddress:"127.0.0.1",tags:{"kuma.io/service":"example","kuma.io/protocol":"tcp"}}]}},E={name:"AddNewServicesCode",components:{CodeBlock:v,OnboardingNavigation:C,OnboardingHeading:y,OnboardingPage:O,LoadingBox:x,KCard:b},data(){return{productName:f,githubLink:"https://github.com/kumahq/kuma-counter-demo/",githubLinkReadme:"https://github.com/kumahq/kuma-counter-demo/blob/master/README.md",k8sRunCommand:"kubectl apply -f https://bit.ly/3Kh2Try",generateDpTokenCode:"kumactl generate dataplane-token --name=redis > kuma-token-redis",startDpCode:`kuma-dp run \\
  --cp-address=${D()} \\
  --dataplane=${`"${k(R)}"`} \\
  --dataplane-token-file=kuma-token-example`,hasDPPs:!1,DPPsTimeout:null}},computed:{..._({environment:"config/getEnvironment"}),isKubernetes(){return this.environment==="kubernetes"}},created(){this.getDPPs()},unmounted(){clearTimeout(this.DPPsTimeout)},methods:{async getDPPs(){try{const{total:i}=await P.getAllDataplanes();this.hasDPPs=i>0}catch(i){console.error(i)}this.hasDPPs||(this.DPPsTimeout=setTimeout(()=>this.getDPPs(),B))}}},K=s("p",{class:"text-center mb-12"},`
        The demo application includes two services: a Redis backend to store a counter value,
        and a frontend web UI to show and increment the counter.
      `,-1),S=s("p",null,"To run execute the following command:",-1),V={key:1},G=s("p",null,"Clone the GitHub repository for the demo application:",-1),H=["href"],j={class:"text-center my-4"},I={key:0,class:"text-green-500","data-testid":"dpps-connected"},M={key:1,class:"text-red-500","data-testid":"dpps-disconnected"},U={key:0,class:"flex justify-center"};function q(i,F,z,J,e,m){const l=t("OnboardingHeading"),c=t("CodeBlock"),p=t("KCard"),u=t("LoadingBox"),h=t("OnboardingNavigation"),g=t("OnboardingPage");return n(),N(g,null,{header:d(()=>[a(l,{title:"Add services"})]),content:d(()=>[K,o(),m.isKubernetes?(n(),r(T,{key:0},[S,o(),a(c,{id:"code-block-kubernetes-command",language:"bash",code:e.k8sRunCommand},null,8,["code"])],64)):(n(),r("div",V,[G,o(),a(c,{id:"code-block-clone-command",language:"bash",code:e.githubLink},null,8,["code"]),o(),a(p,{title:"And follow the instructions in the README","border-variant":"noBorder"},{body:d(()=>[s("a",{target:"_blank",class:"external-link-code-block",href:e.githubLinkReadme},w(e.githubLinkReadme),9,H)]),_:1})])),o(),s("div",null,[s("p",j,[o(`
          DPPs status:
          `),e.hasDPPs?(n(),r("span",I,"Connected")):(n(),r("span",M,"Disconeccted"))]),o(),e.hasDPPs?L("",!0):(n(),r("div",U,[a(u)]))])]),navigation:d(()=>[a(h,{"next-step":"onboarding-dataplanes-overview","previous-step":"onboarding-add-services","should-allow-next":e.hasDPPs},null,8,["should-allow-next"])]),_:1})}const de=A(E,[["render",q]]);export{de as default};
