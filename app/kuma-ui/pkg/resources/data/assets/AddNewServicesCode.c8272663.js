import{_ as g,E as _,$ as k,i as b,K as P,c as f,w as r,r as t,o,a as n,e as s,F as v,f as a,t as D,b as x,k as C}from"./index.dbfc69fe.js";import{j as y}from"./index.58caa11d.js";import{k as O}from"./configUrl.b3f5cf10.js";import{C as w}from"./CodeBlock.83ca39be.js";import{L as N}from"./LoadingBox.fcadec95.js";import{O as L}from"./OnboardingNavigation.bf337693.js";import{O as T,a as B}from"./OnboardingPage.8d6a6242.js";import"./_commonjsHelpers.712cc82f.js";const A=1e3,R={name:"AddNewServicesCode",components:{CodeBlock:w,OnboardingNavigation:L,OnboardingHeading:T,OnboardingPage:B,LoadingBox:N,KCard:_},data(){return{productName:k,githubLink:"https://github.com/kumahq/kuma-counter-demo/",githubLinkReadme:"https://github.com/kumahq/kuma-counter-demo/blob/master/README.md",k8sRunCommand:"kubectl apply -f https://bit.ly/3Kh2Try",generateDpTokenCode:"kumactl generate dataplane-token --name=redis > kuma-token-redis",startDpCode:`kuma-dp run \\
      --cp-address=${O()} \\
      --dataplane=${`"${y({type:"Dataplane",mesh:"default",name:"example",networking:{address:"localhost",inbound:[{port:7777,servicePort:7777,serviceAddress:"127.0.0.1",tags:{"kuma.io/service":"example","kuma.io/protocol":"tcp"}}]}})}"`} \\
      --dataplane-token-file=kuma-token-example`,hasDPPs:!1,DPPsTimeout:null}},computed:{...b({environment:"config/getEnvironment"}),isKubernetes(){return this.environment==="kubernetes"}},created(){this.getDPPs()},unmounted(){clearTimeout(this.DPPsTimeout)},methods:{async getDPPs(){try{const{total:i}=await P.getAllDataplanes();this.hasDPPs=i>0}catch(i){console.error(i)}this.hasDPPs||(this.DPPsTimeout=setTimeout(()=>this.getDPPs(),A))}}},E=a("p",{class:"text-center mb-12"}," The demo application includes two services: a Redis backend to store a counter value, and a frontend web UI to show and increment the counter. ",-1),K=a("p",null,"To run execute the following command:",-1),V={key:1},G=a("p",null,"Clone the GitHub repository for the demo application:",-1),H=["href"],S={class:"text-center my-4"},j={key:0,class:"text-green-500","data-testid":"dpps-connected"},I={key:1,class:"text-red-500","data-testid":"dpps-disconnected"},M={key:0,class:"flex justify-center"};function U(i,q,F,z,e,c){const l=t("OnboardingHeading"),d=t("CodeBlock"),m=t("KCard"),u=t("LoadingBox"),p=t("OnboardingNavigation"),h=t("OnboardingPage");return o(),f(h,null,{header:r(()=>[n(l,{title:"Add services"})]),content:r(()=>[E,c.isKubernetes?(o(),s(v,{key:0},[K,n(d,{language:"bash",code:e.k8sRunCommand},null,8,["code"])],64)):(o(),s("div",V,[G,n(d,{language:"bash",code:e.githubLink},null,8,["code"]),n(m,{title:"And follow the instructions in the README","border-variant":"noBorder"},{body:r(()=>[a("a",{target:"_blank",class:"external-link-code-block",href:e.githubLinkReadme},D(e.githubLinkReadme),9,H)]),_:1})])),a("div",null,[a("p",S,[x(" DPPs status: "),e.hasDPPs?(o(),s("span",j,"Connected")):(o(),s("span",I,"Disconeccted"))]),e.hasDPPs?C("",!0):(o(),s("div",M,[n(u)]))])]),navigation:r(()=>[n(p,{"next-step":"onboarding-dataplanes-overview","previous-step":"onboarding-add-services","should-allow-next":e.hasDPPs},null,8,["should-allow-next"])]),_:1})}const te=g(R,[["render",U]]);export{te as default};
